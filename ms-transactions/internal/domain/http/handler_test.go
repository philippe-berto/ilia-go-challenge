package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"transactions/internal/domain/repository"
	"transactions/internal/domain/transaction"
	"transactions/internal/dto"
	"transactions/internal/utils/middleware"
	"transactions/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	testUserID = "user-abc-123"
	testTxID   = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
)

func newTestRouter(t *testing.T) (chi.Router, *mocks.MockService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	svc := mocks.NewMockService(ctrl)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	r := chi.NewRouter()
	// inject userID directly, bypassing JWT middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), middleware.UserIDKey, testUserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	h := &handler{s: svc, logger: logger}
	r.Post("/transactions", h.createTransaction)
	r.Get("/transactions", h.getTransactions)
	r.Get("/balance", h.getBalance)

	return r, svc
}

func TestHandler_CreateTransaction(t *testing.T) {
	t.Run("Should create transaction and return 200", func(t *testing.T) {
		r, svc := newTestRouter(t)

		expected := &dto.TransactionOutput{
			ID:     testTxID,
			UserID: testUserID,
			Type:   "CREDIT",
			Amount: 100.00,
		}

		svc.EXPECT().
			CreateTransaction(gomock.Any(), &transaction.Transaction{
				ID:     testTxID,
				UserID: testUserID,
				Type:   transaction.Credit,
				Amount: 100.00,
			}).
			Return(expected, nil)

		body, _ := json.Marshal(map[string]any{"id": testTxID, "type": "CREDIT", "amount": 100.00})
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var out dto.TransactionOutput
		require.NoError(t, json.NewDecoder(w.Body).Decode(&out))
		assert.Equal(t, expected.ID, out.ID)
		assert.Equal(t, expected.Amount, out.Amount)
	})

	t.Run("Should return 400 when body is invalid", func(t *testing.T) {
		r, _ := newTestRouter(t)

		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString("not-json"))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Should return 400 when transaction id is not a valid UUID v4", func(t *testing.T) {
		r, _ := newTestRouter(t)

		body, _ := json.Marshal(map[string]any{"id": "not-a-uuid", "type": "CREDIT", "amount": 50.00})
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Should return 400 when transaction type is invalid", func(t *testing.T) {
		r, _ := newTestRouter(t)

		body, _ := json.Marshal(map[string]any{"id": testTxID, "type": "UNKNOWN", "amount": 50.00})
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Should return 422 when balance is insufficient", func(t *testing.T) {
		r, svc := newTestRouter(t)

		svc.EXPECT().
			CreateTransaction(gomock.Any(), gomock.Any()).
			Return(nil, repository.ErrInsufficientBalance)

		body, _ := json.Marshal(map[string]any{"id": testTxID, "type": "DEBIT", "amount": 999.00})
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Should return 500 when service returns unexpected error", func(t *testing.T) {
		r, svc := newTestRouter(t)

		svc.EXPECT().
			CreateTransaction(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("db connection lost"))

		body, _ := json.Marshal(map[string]any{"id": testTxID, "type": "DEBIT", "amount": 999.00})
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_GetTransactions(t *testing.T) {
	t.Run("Should return all transactions when no type filter", func(t *testing.T) {
		r, svc := newTestRouter(t)

		expected := []*dto.TransactionOutput{
			{ID: "tx-1", UserID: testUserID, Type: "CREDIT", Amount: 100.00},
			{ID: "tx-2", UserID: testUserID, Type: "DEBIT", Amount: 30.00},
		}

		svc.EXPECT().
			GetTransactionsByUser(gomock.Any(), testUserID).
			Return(expected, nil)

		req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var out []*dto.TransactionOutput
		require.NoError(t, json.NewDecoder(w.Body).Decode(&out))
		assert.Len(t, out, len(expected))
	})

	t.Run("Should return transactions filtered by type", func(t *testing.T) {
		r, svc := newTestRouter(t)

		expected := []*dto.TransactionOutput{
			{ID: "tx-1", UserID: testUserID, Type: "CREDIT", Amount: 100.00},
		}

		svc.EXPECT().
			GetTransactionsByType(gomock.Any(), testUserID, "CREDIT").
			Return(expected, nil)

		req := httptest.NewRequest(http.MethodGet, "/transactions?type=CREDIT", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var out []*dto.TransactionOutput
		require.NoError(t, json.NewDecoder(w.Body).Decode(&out))
		assert.Len(t, out, 1)
	})
}

func TestHandler_GetBalance(t *testing.T) {
	t.Run("Should return balance with 200", func(t *testing.T) {
		r, svc := newTestRouter(t)

		svc.EXPECT().
			GetBalance(gomock.Any(), testUserID).
			Return(150.00, nil)

		req := httptest.NewRequest(http.MethodGet, "/balance", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var out map[string]float64
		require.NoError(t, json.NewDecoder(w.Body).Decode(&out))
		assert.Equal(t, 150.00, out["amount"])
	})

	t.Run("Should return 500 when service returns error", func(t *testing.T) {
		r, svc := newTestRouter(t)

		svc.EXPECT().
			GetBalance(gomock.Any(), testUserID).
			Return(float64(0), errors.New("db error"))

		req := httptest.NewRequest(http.MethodGet, "/balance", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
