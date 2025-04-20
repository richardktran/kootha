package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/service/quizsession"
	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
)

type RoomHandler struct {
	service *quizsession.Service
}

func NewRoomHandler(service *quizsession.Service) *RoomHandler {
	return &RoomHandler{
		service: service,
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

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received create room request")

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

	log.Printf("Creating room with name: %s, hostId: %s", req.Name, req.HostID)

	quizSession := &model.QuizSession{
		Name: req.Name,
		ID:   req.HostID,
	}

	session, err := h.service.CreateQuizSession(r.Context(), quizSession)
	if err != nil {
		log.Printf("Error creating quiz session: %v", err)
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}

	log.Printf("Room created successfully: %+v", session)

	response := CreateRoomResponse{
		RoomID: session.ID,
		Name:   session.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

type JoinRoomRequest struct {
	RoomID string `json:"roomId"`
	UserID string `json:"userId"`
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

	session, err := h.service.JoinQuiz(r.Context(), req.RoomID, req.UserID)
	if err != nil {
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
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

	session, err := h.service.GetSessionById(r.Context(), roomID)
	if err != nil {
		http.Error(w, "Failed to get room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}
