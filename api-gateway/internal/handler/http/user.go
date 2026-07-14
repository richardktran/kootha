package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/richardktran/realtime-quiz/user-service/pkg/model"
)

type userGateway interface {
	CreateUser(ctx context.Context, name string) (*model.User, error)
}

type UserHandler struct {
	gateway userGateway
}

func NewUserHandler(gateway userGateway) *UserHandler {
	return &UserHandler{
		gateway: gateway,
	}
}

type CreateUserRequest struct {
	Name string `json:"name"`
}

type CreateUserResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	user, err := h.gateway.CreateUser(r.Context(), req.Name)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	response := CreateUserResponse{
		ID:   user.ID,
		Name: user.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
