package ws

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/richardktran/realtime-quiz/gen"
	quizsessionGW "github.com/richardktran/realtime-quiz/notification-service/internal/gateway/quizsession"
	"github.com/richardktran/realtime-quiz/notification-service/internal/fanout"
)

type Server struct {
	upgrader websocket.Upgrader
	hub      *Hub
	quizGW   *quizsessionGW.Gateway
	fanout   *fanout.Publisher
}

type clientParticipant struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type clientRoom struct {
	ID                   string              `json:"id"`
	Name                 string              `json:"name"`
	HostID               string              `json:"hostId"`
	Participants         []clientParticipant `json:"participants"`
	Status               string              `json:"status"`
	CurrentQuestionIndex int                 `json:"currentQuestionIndex"`
	Questions            []interface{}       `json:"questions"`
}

func NewServer(hub *Hub, quizGW *quizsessionGW.Gateway, fanout *fanout.Publisher) *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				return origin == "http://localhost:3000" || origin == "http://127.0.0.1:3000"
			},
			ReadBufferSize:    8192,
			WriteBufferSize:   8192,
			EnableCompression: true,
			HandshakeTimeout:  10 * time.Second,
		},
		hub:    hub,
		quizGW: quizGW,
		fanout: fanout,
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("new WebSocket connection: %v", conn.RemoteAddr())

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			s.handleDisconnect(conn)
			break
		}

		switch msg.Type {
		case "JOIN_ROOM":
			s.handleJoinRoom(conn, msg.Payload)
		case "START_QUIZ":
			s.handleStartQuiz(conn, msg.Payload)
		case "SUBMIT_ANSWER":
			s.handleSubmitAnswer(conn, msg.Payload)
		case "NEXT_QUESTION":
			s.handleNextQuestion(conn, msg.Payload)
		}
	}
}

func (s *Server) handleJoinRoom(conn *websocket.Conn, data map[string]interface{}) {
	roomID, ok := data["roomId"].(string)
	if !ok {
		s.sendError(conn, "invalid room ID")
		return
	}

	userID, ok := data["userId"].(string)
	if !ok {
		s.sendError(conn, "invalid user ID")
		return
	}

	name, _ := data["name"].(string)
	if name == "" {
		name = userID
	}

	ctx := context.Background()
	resp, err := s.quizGW.JoinQuiz(ctx, roomID, userID, name)
	if err != nil {
		log.Printf("JoinQuiz failed: %v", err)
		s.sendError(conn, "failed to join room")
		return
	}

	session := resp.GetQuizSession()
	s.hub.Join(roomID, conn, &ClientInfo{UserID: userID, Name: name}, session.GetHostId())
	s.sendToConn(conn, "room_joined", protoToClientRoom(session))
}

func (s *Server) handleStartQuiz(conn *websocket.Conn, data map[string]interface{}) {
	roomID, ok := data["roomId"].(string)
	if !ok {
		s.sendError(conn, "invalid room ID")
		return
	}

	userID, _ := data["userId"].(string)
	if userID == "" {
		if info, _ := s.hub.GetClient(conn); info != nil {
			userID = info.UserID
		}
	}
	if userID == "" {
		s.sendError(conn, "invalid user ID")
		return
	}

	ctx := context.Background()
	resp, err := s.quizGW.StartSession(ctx, roomID, userID, 5)
	if err != nil {
		log.Printf("StartSession failed: %v", err)
		s.sendError(conn, "failed to start quiz")
		return
	}

	// Push immediately so clients don't depend solely on the Kafka consumer path.
	q := resp.GetQuestion()
	if q != nil {
		questionIndex := int32(0)
		if session := resp.GetQuizSession(); session != nil {
			questionIndex = session.GetCurrentIndex()
		}
		if err := s.fanout.Publish(ctx, roomID, "question_started", map[string]interface{}{
			"questionIndex": questionIndex,
			"question": map[string]interface{}{
				"id":        q.GetId(),
				"question":  q.GetQuestion(),
				"options":   q.GetOptions(),
				"timeLimit": q.GetTimeLimit(),
			},
			"status": "in-progress",
		}); err != nil {
			log.Printf("failed to fanout question_started: %v", err)
		}
	}
}

func (s *Server) handleNextQuestion(conn *websocket.Conn, data map[string]interface{}) {
	roomID, ok := data["roomId"].(string)
	if !ok {
		s.sendError(conn, "invalid room ID")
		return
	}

	userID, _ := data["userId"].(string)
	if userID == "" {
		if info, _ := s.hub.GetClient(conn); info != nil {
			userID = info.UserID
		}
	}
	if userID == "" {
		s.sendError(conn, "invalid user ID")
		return
	}

	ctx := context.Background()
	resp, err := s.quizGW.NextQuestion(ctx, roomID, userID)
	if err != nil {
		log.Printf("NextQuestion failed: %v", err)
		s.sendError(conn, "failed to advance question")
		return
	}

	if resp.GetFinished() {
		return
	}

	q := resp.GetQuestion()
	if q != nil {
		if err := s.fanout.Publish(ctx, roomID, "question_started", map[string]interface{}{
			"questionIndex": resp.GetQuestionIndex(),
			"question": map[string]interface{}{
				"id":        q.GetId(),
				"question":  q.GetQuestion(),
				"options":   q.GetOptions(),
				"timeLimit": q.GetTimeLimit(),
			},
		}); err != nil {
			log.Printf("failed to fanout question_started: %v", err)
		}
	}
}

func (s *Server) handleSubmitAnswer(conn *websocket.Conn, data map[string]interface{}) {
	roomID, ok := data["roomId"].(string)
	if !ok {
		s.sendError(conn, "invalid room ID")
		return
	}

	userID, _ := data["userId"].(string)
	if userID == "" {
		if info, _ := s.hub.GetClient(conn); info != nil {
			userID = info.UserID
		}
	}
	if userID == "" {
		s.sendError(conn, "invalid user ID")
		return
	}

	answer, ok := data["answer"].(map[string]interface{})
	if !ok {
		s.sendError(conn, "invalid answer")
		return
	}

	questionID, ok := answer["questionId"].(string)
	if !ok {
		s.sendError(conn, "invalid question ID")
		return
	}

	selectedOption, ok := int32FromJSON(answer["selectedOption"])
	if !ok {
		s.sendError(conn, "invalid selected option")
		return
	}

	timeToAnswer, ok := int32FromJSON(answer["timeToAnswer"])
	if !ok {
		timeToAnswer = 0
	}

	ctx := context.Background()
	if _, err := s.quizGW.SubmitAnswer(ctx, roomID, userID, questionID, selectedOption, timeToAnswer); err != nil {
		log.Printf("SubmitAnswer failed: %v", err)
		s.sendError(conn, "failed to submit answer")
	}
}

func (s *Server) handleDisconnect(conn *websocket.Conn) {
	sessionID, userID, wasHost := s.hub.Leave(conn)
	if sessionID == "" || userID == "" {
		return
	}

	ctx := context.Background()
	if wasHost {
		if resp, err := s.quizGW.ReassignHost(ctx, sessionID, userID); err != nil {
			log.Printf("ReassignHost failed: %v", err)
		} else if resp.GetHostId() != "" {
			s.hub.SetHost(sessionID, resp.GetHostId())
		}
	}

	if err := s.fanout.Publish(ctx, sessionID, "participant_left", map[string]interface{}{
		"participantId": userID,
	}); err != nil {
		log.Printf("failed to fanout participant_left: %v", err)
	}
}

func (s *Server) sendToConn(conn *websocket.Conn, eventType string, data interface{}) {
	msg := Message{
		Type:    eventType,
		Payload: map[string]interface{}{"data": data},
	}
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("failed to send message: %v", err)
	}
}

func (s *Server) sendError(conn *websocket.Conn, message string) {
	msg := Message{
		Type:    "error",
		Payload: map[string]interface{}{"message": message},
	}
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("failed to send error: %v", err)
	}
}

func protoToClientRoom(session *gen.QuizSession) clientRoom {
	participants := make([]clientParticipant, 0, len(session.GetParticipants()))
	for _, p := range session.GetParticipants() {
		participants = append(participants, clientParticipant{
			ID:    p.GetId(),
			Name:  p.GetName(),
			Score: int(p.GetScore()),
		})
	}

	return clientRoom{
		ID:                   session.GetId(),
		Name:                 session.GetName(),
		HostID:               session.GetHostId(),
		Participants:         participants,
		Status:               session.GetStatus(),
		CurrentQuestionIndex: int(session.GetCurrentIndex()),
		Questions:            []interface{}{},
	}
}

func int32FromJSON(v interface{}) (int32, bool) {
	switch n := v.(type) {
	case float64:
		return int32(n), true
	case int:
		return int32(n), true
	case int32:
		return n, true
	case int64:
		return int32(n), true
	default:
		return 0, false
	}
}
