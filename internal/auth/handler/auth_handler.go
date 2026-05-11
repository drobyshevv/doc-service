package handler

import (
	"encoding/json"
	"net/http"

	"github.com/drobyshevv/doc-service/internal/auth/model"
	"github.com/drobyshevv/doc-service/internal/auth/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, tokens, err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"user": map[string]any{
			"id":    user.ID.String(),
			"email": user.Email,
			"role":  user.Role,
		},
		"tokens": tokens,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, tokens, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"user": map[string]any{
			"id":    user.ID.String(),
			"email": user.Email,
			"role":  user.Role,
		},
		"tokens": tokens,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, tokens, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	resp := map[string]any{
		"user": map[string]any{
			"id":    user.ID.String(),
			"email": user.Email,
			"role":  user.Role,
		},
		"tokens": tokens,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
