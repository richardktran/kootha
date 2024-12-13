package http

import (
	"log"
	"net/http"
)

type Handler struct {
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Get user")
}
