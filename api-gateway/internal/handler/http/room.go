package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/richardktran/realtime-quiz/gen"
)

type quizSessionGateway interface {
	CreateQuizSession(ctx context.Context, name, hostID string) (*gen.QuizSession, error)
	GetQuizSessionById(ctx context.Context, id string) (*gen.QuizSession, error)
	JoinQuiz(ctx context.Context, quizSessionID, userID, name string) (*gen.QuizSession, error)
}

type RoomHandler struct {
	gateway quizSessionGateway
}

func NewRoomHandler(gateway quizSessionGateway) *RoomHandler {
	return &RoomHandler{
		gateway: gateway,
	}
}

type CreateRoomRequest struct {
	Name   string `json:"name"`
	HostID string `json:"hostId"`
}

type CreateRoomResponse struct {
	RoomID string `json:"roomId"`
	Name   string `json:"name"`
}

type JoinRoomRequest struct {
	RoomID string `json:"roomId"`
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

type ParticipantResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score int32  `json:"score"`
}

type RoomResponse struct {
	ID                   string                `json:"id"`
	Name                 string                `json:"name"`
	HostID               string                `json:"hostId"`
	Status               string                `json:"status"`
	Participants         []ParticipantResponse `json:"participants"`
	Duration             int32                 `json:"duration"`
	CurrentQuestionIndex int32                 `json:"currentQuestionIndex"`
}

func roomFromProto(session *gen.QuizSession) RoomResponse {
	participants := make([]ParticipantResponse, 0, len(session.GetParticipants()))
	for _, p := range session.GetParticipants() {
		participants = append(participants, ParticipantResponse{
			ID:    p.GetId(),
			Name:  p.GetName(),
			Score: p.GetScore(),
		})
	}

	return RoomResponse{
		ID:                   session.GetId(),
		Name:                 session.GetName(),
		HostID:               session.GetHostId(),
		Status:               session.GetStatus(),
		Participants:         participants,
		Duration:             session.GetDuration(),
		CurrentQuestionIndex: session.GetCurrentIndex(),
	}
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.HostID == "" {
		http.Error(w, "Name and hostId are required", http.StatusBadRequest)
		return
	}

	session, err := h.gateway.CreateQuizSession(r.Context(), req.Name, req.HostID)
	if err != nil {
		log.Printf("Error creating quiz session: %v", err)
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}

	response := CreateRoomResponse{
		RoomID: session.GetId(),
		Name:   session.GetName(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *RoomHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomID := r.URL.Query().Get("id")
	if roomID == "" {
		http.Error(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	session, err := h.gateway.GetQuizSessionById(r.Context(), roomID)
	if err != nil {
		log.Printf("Error getting quiz session: %v", err)
		http.Error(w, "Failed to get room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(roomFromProto(session)); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RoomID == "" || req.UserID == "" || req.Name == "" {
		http.Error(w, "roomId, userId, and name are required", http.StatusBadRequest)
		return
	}

	session, err := h.gateway.JoinQuiz(r.Context(), req.RoomID, req.UserID, req.Name)
	if err != nil {
		log.Printf("Error joining quiz session: %v", err)
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(roomFromProto(session)); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
