package handler

import (
	"encoding/json"
	"net/http"

	"github.com/drobyshevv/doc-service/internal/auth/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	resp := map[string]any{
		"id":    user.ID.String(),
		"email": user.Email,
		"role":  user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
