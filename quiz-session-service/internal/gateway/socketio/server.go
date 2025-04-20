package socketio

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/service/quizsession"
	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
)

type Server struct {
	upgrader websocket.Upgrader
	service  *quizsession.Service
	rooms    sync.Map // map[string]*Room
	clients  sync.Map // map[*websocket.Conn]string // conn -> roomID
}

type Room struct {
	ID                   string
	HostID               string
	Participants         map[string]*model.Participant
	Status               string
	CurrentQuestionIndex int
	Questions            []model.Question
	Connections          map[*websocket.Conn]string // conn -> participantID
}

type Message struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

func NewServer(service *quizsession.Service) *Server {
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
		service: service,
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers for WebSocket upgrade
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("New WebSocket connection: %v", conn.RemoteAddr())

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
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
	log.Printf("Received join_room event with data: %+v", data)

	roomID, ok := data["roomId"].(string)
	if !ok {
		s.sendError(conn, "Invalid room ID")
		return
	}

	userID, ok := data["userId"].(string)
	if !ok {
		s.sendError(conn, "Invalid user ID")
		return
	}

	// Create participant
	participant := &model.Participant{
		ID:    userID,
		Name:  data["name"].(string),
		Score: 0,
	}

	// Get or create room
	roomValue, loaded := s.rooms.LoadOrStore(roomID, &Room{
		ID:           roomID,
		Participants: make(map[string]*model.Participant),
		Status:       "waiting",
		Connections:  make(map[*websocket.Conn]string),
	})
	room := roomValue.(*Room)

	// Add connection and participant to room
	room.Connections[conn] = userID
	room.Participants[userID] = participant
	s.clients.Store(conn, roomID)

	// If this is the first participant, make them the host
	if !loaded || len(room.Participants) == 1 {
		room.HostID = userID
	}

	// Send room joined event to the participant
	s.sendToConn(conn, "room_joined", room)

	// Broadcast participant joined event to room
	s.broadcastToRoom(roomID, "participant_joined", map[string]interface{}{
		"participant": participant,
	}, conn)
}

func (s *Server) handleStartQuiz(conn *websocket.Conn, data map[string]interface{}) {
	roomID := data["roomId"].(string)
	roomValue, ok := s.rooms.Load(roomID)
	if !ok {
		return
	}
	room := roomValue.(*Room)

	// Check if the sender is the host
	if room.Connections[conn] != room.HostID {
		return
	}

	// Update room status
	room.Status = "in-progress"
	room.CurrentQuestionIndex = 0

	// Broadcast quiz started event
	s.broadcastToRoom(roomID, "question_started", map[string]interface{}{
		"questionIndex": 0,
		"question":      room.Questions[0],
	}, nil)
}

func (s *Server) handleSubmitAnswer(conn *websocket.Conn, data map[string]interface{}) {
	roomID := data["roomId"].(string)
	answer := data["answer"].(map[string]interface{})
	roomValue, ok := s.rooms.Load(roomID)
	if !ok {
		return
	}
	room := roomValue.(*Room)

	participantID := room.Connections[conn]
	participant := room.Participants[participantID]
	question := room.Questions[room.CurrentQuestionIndex]
	if answer["selectedOption"].(int) == question.CorrectAnswer {
		participant.Score++
	}

	// Broadcast leaderboard update
	s.broadcastToRoom(roomID, "leaderboard_update", map[string]interface{}{
		"participants": room.Participants,
	}, nil)
}

func (s *Server) handleNextQuestion(conn *websocket.Conn, data map[string]interface{}) {
	roomID := data["roomId"].(string)
	roomValue, ok := s.rooms.Load(roomID)
	if !ok {
		return
	}
	room := roomValue.(*Room)

	// Check if the sender is the host
	if room.Connections[conn] != room.HostID {
		return
	}

	// Move to next question
	room.CurrentQuestionIndex++
	if room.CurrentQuestionIndex >= len(room.Questions) {
		// Quiz ended
		room.Status = "finished"
		s.broadcastToRoom(roomID, "quiz_ended", map[string]interface{}{
			"participants": room.Participants,
		}, nil)
		return
	}

	// Broadcast next question
	s.broadcastToRoom(roomID, "question_started", map[string]interface{}{
		"questionIndex": room.CurrentQuestionIndex,
		"question":      room.Questions[room.CurrentQuestionIndex],
	}, nil)
}

func (s *Server) handleDisconnect(conn *websocket.Conn) {
	if roomID, ok := s.clients.LoadAndDelete(conn); ok {
		if roomValue, ok := s.rooms.Load(roomID); ok {
			room := roomValue.(*Room)
			participantID := room.Connections[conn]
			delete(room.Connections, conn)
			delete(room.Participants, participantID)

			// If host left, assign new host
			if room.HostID == participantID && len(room.Participants) > 0 {
				for _, pid := range room.Connections {
					room.HostID = pid
					break
				}
			}

			// Broadcast participant left event
			s.broadcastToRoom(roomID.(string), "participant_left", map[string]interface{}{
				"participantId": participantID,
			}, nil)
		}
	}
}

func (s *Server) sendToConn(conn *websocket.Conn, eventType string, data interface{}) {
	msg := Message{
		Type:    eventType,
		Payload: map[string]interface{}{"data": data},
	}
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func (s *Server) sendError(conn *websocket.Conn, message string) {
	s.sendToConn(conn, "error", map[string]string{"message": message})
}

func (s *Server) broadcastToRoom(roomID string, eventType string, data interface{}, exclude *websocket.Conn) {
	if roomValue, ok := s.rooms.Load(roomID); ok {
		room := roomValue.(*Room)
		msg := Message{
			Type:    eventType,
			Payload: map[string]interface{}{"data": data},
		}
		for conn := range room.Connections {
			if conn != exclude {
				if err := conn.WriteJSON(msg); err != nil {
					log.Printf("Failed to broadcast message: %v", err)
				}
			}
		}
	}
}

func (s *Server) Start(ctx context.Context) error {
	log.Println("Starting WebSocket server...")
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return nil
}

func (s *Server) GetServer() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.HandleWebSocket(w, r)
	})
}
