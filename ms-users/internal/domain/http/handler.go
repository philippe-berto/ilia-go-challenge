package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"users/internal/domain/service"
	"users/internal/dto"
	"users/internal/utils/jwt"
	"users/internal/utils/middleware"

	"github.com/go-chi/chi/v5"
)

type (
	Service interface {
		CreateUser(ctx context.Context, firstName, lastName, email, password string) (*dto.UserOutput, error)
		GetUserByID(ctx context.Context, id string) (*dto.UserOutput, error)
		GetUsers(ctx context.Context) ([]*dto.UserOutput, error)
		UpdateUser(ctx context.Context, id, firstName, lastName, email string) (*dto.UserOutput, error)
		DeleteUser(ctx context.Context, id string) error
		Authenticate(ctx context.Context, email, password string) (*dto.AuthOutput, error)
	}

	handler struct {
		s      Service
		logger *slog.Logger
	}
)

func Register(router chi.Router, svc Service, logger *slog.Logger, jwtClient *jwt.Client) {
	h := &handler{s: svc, logger: logger}

	// Public routes
	router.Post("/users", h.createUser)
	router.Post("/auth", h.authenticate)

	// Protected routes
	router.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwtClient))
		r.Get("/users", h.getUsers)
		r.Get("/users/{id}", h.getUserByID)
		r.Patch("/users/{id}", h.updateUser)
		r.Delete("/users/{id}", h.deleteUser)
	})
}

func (h *handler) createUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	output, err := h.s.CreateUser(r.Context(), input.FirstName, input.LastName, input.Email, input.Password)
	if err != nil {
		h.logger.Error("failed to create user", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(output)
}

func (h *handler) authenticate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		User     *struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if input.User != nil {
		input.Email = input.User.Email
		input.Password = input.User.Password
	}

	output, err := h.s.Authenticate(r.Context(), input.Email, input.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "invalid email or password", http.StatusUnauthorized)
			return
		}
		h.logger.Error("failed to authenticate", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func (h *handler) getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.s.GetUsers(r.Context())
	if err != nil {
		h.logger.Error("failed to get users", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *handler) getUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	output, err := h.s.GetUserByID(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get user", "error", err)
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func (h *handler) updateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	output, err := h.s.UpdateUser(r.Context(), id, input.FirstName, input.LastName, input.Email)
	if err != nil {
		h.logger.Error("failed to update user", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func (h *handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.s.DeleteUser(r.Context(), id); err != nil {
		h.logger.Error("failed to delete user", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
