package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"transactions/internal/domain/transaction"
	"transactions/internal/dto"
	"transactions/internal/utils/jwt"
	"transactions/internal/utils/middleware"

	"github.com/go-chi/chi/v5"
)

type (
	Service interface {
		CreateTransaction(ctx context.Context, transaction *transaction.Transaction) (*dto.TransactionOutput, error)
		GetTransactionsByUser(ctx context.Context, userID string) ([]*dto.TransactionOutput, error)
		GetTransactionsByType(ctx context.Context, userID, transactionType string) ([]*dto.TransactionOutput, error)
		GetBalance(ctx context.Context, userID string) (float64, error)
	}
	handler struct {
		s      Service
		logger *slog.Logger
	}
)

func Register(router chi.Router, service Service, logger *slog.Logger, jwtClient *jwt.Client) {
	h := &handler{s: service, logger: logger}

	router.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwtClient))
		r.Post("/transactions", h.createTransaction)
		r.Get("/transactions", h.getTransactions)
		r.Get("/balance", h.getBalance)
	})
}

func (h *handler) createTransaction(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID     string  `json:"id"`
		Type   string  `json:"type"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	t, err := transaction.New(input.ID, userID, transaction.TransactionType(strings.ToLower(input.Type)), input.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	output, err := h.s.CreateTransaction(r.Context(), t)
	if err != nil {
		h.logger.Error("failed to create transaction", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func (h *handler) getTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var result []*dto.TransactionOutput
	var err error

	if transactionType := r.URL.Query().Get("type"); transactionType != "" {
		result, err = h.s.GetTransactionsByType(r.Context(), userID, transactionType)
	} else {
		result, err = h.s.GetTransactionsByUser(r.Context(), userID)
	}

	if err != nil {
		h.logger.Error("failed to get transactions", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *handler) getBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.s.GetBalance(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get balance", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"balance": balance})
}
