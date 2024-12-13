package http

import (
	"fmt"
	"net/http"

	idgeneration "github.com/richardktran/realtime-quiz/id-generation-service/internal/service/idGeneration"
)

type Handler struct {
	service *idgeneration.Service
}

func New(svc *idgeneration.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

func (h *Handler) GenerateId(w http.ResponseWriter, r *http.Request) {
	entity := r.URL.Query().Get("entity")

	ctx := r.Context()
	id := h.service.GenerateId(ctx, entity)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"id": "%s"}`, id)))
}
