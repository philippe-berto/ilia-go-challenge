package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"users/internal/domain/service"
	"users/internal/dto"
	"users/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestRouter(t *testing.T) (chi.Router, *mocks.MockService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	svc := mocks.NewMockService(ctrl)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	r := chi.NewRouter()
	h := &handler{s: svc, logger: logger}
	r.Post("/users", h.createUser)
	r.Post("/auth", h.authenticate)
	r.Get("/users", h.getUsers)
	r.Get("/users/{id}", h.getUserByID)
	r.Patch("/users/{id}", h.updateUser)
	r.Delete("/users/{id}", h.deleteUser)

	return r, svc
}

func TestHandler_CreateUser(t *testing.T) {
	t.Run("Should create user and return 201", func(t *testing.T) {
		r, svc := newTestRouter(t)

		expected := &dto.UserOutput{
			ID:        "user-1",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		}

		svc.EXPECT().
			CreateUser(gomock.Any(), "John", "Doe", "john@example.com", "secret123").
			Return(expected, nil)

		body, _ := json.Marshal(map[string]any{
			"first_name": "John",
			"last_name":  "Doe",
			"email":      "john@example.com",
			"password":   "secret123",
		})
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		var out dto.UserOutput
		require.NoError(t, json.NewDecoder(w.Body).Decode(&out))
		assert.Equal(t, expected.ID, out.ID)
		assert.Equal(t, expected.Email, out.Email)
	})

	t.Run("Should return 400 when body is invalid", func(t *testing.T) {
		r, _ := newTestRouter(t)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("not-json"))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Should return 400 when service returns error", func(t *testing.T) {
		r, svc := newTestRouter(t)

		svc.EXPECT().
			CreateUser(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errors.New("email already exists"))

		body, _ := json.Marshal(map[string]any{
			"first_name": "John",
			"last_name":  "Doe",
			"email":      "john@example.com",
			"password":   "pass",
		})
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_Authenticate(t *testing.T) {
	t.Run("Should return token on valid credentials", func(t *testing.T) {
		r, svc := newTestRouter(t)

		svc.EXPECT().
			Authenticate(gomock.Any(), "john@example.com", "secret123").
			Return(&dto.AuthOutput{Token: "token-abc"}, nil)

		body, _ := json.Marshal(map[string]any{
			"email":    "john@example.com",
			"password": "secret123",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var out dto.AuthOutput
		require.NoError(t, json.NewDecoder(w.Body).Decode(&out))
		assert.Equal(t, "token-abc", out.Token)
	})

	t.Run("Should return 401 on invalid credentials", func(t *testing.T) {
		r, svc := newTestRouter(t)

		svc.EXPECT().
			Authenticate(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, service.ErrInvalidCredentials)

		body, _ := json.Marshal(map[string]any{"email": "x@x.com", "password": "wrong"})
		req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Should return 400 when body is invalid", func(t *testing.T) {
		r, _ := newTestRouter(t)

		req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewBufferString("not-json"))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_GetUsers(t *testing.T) {
	t.Run("Should return all users", func(t *testing.T) {
		r, svc := newTestRouter(t)

		expected := []*dto.UserOutput{
			{ID: "user-1", FirstName: "John", LastName: "Doe", Email: "john@example.com"},
			{ID: "user-2", FirstName: "Jane", LastName: "Doe", Email: "jane@example.com"},
		}

		svc.EXPECT().
			GetUsers(gomock.Any()).
			Return(expected, nil)

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var out []*dto.UserOutput
		require.NoError(t, json.NewDecoder(w.Body).Decode(&out))
		assert.Len(t, out, len(expected))
	})

	t.Run("Should return user by ID", func(t *testing.T) {
		r, svc := newTestRouter(t)

		expected := &dto.UserOutput{
			ID:        "user-1",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		}

		svc.EXPECT().
			GetUserByID(gomock.Any(), "user-1").
			Return(expected, nil)

		req := httptest.NewRequest(http.MethodGet, "/users/user-1", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var out dto.UserOutput
		require.NoError(t, json.NewDecoder(w.Body).Decode(&out))
		assert.Equal(t, expected.ID, out.ID)
	})

	t.Run("Should return 404 when user is not found", func(t *testing.T) {
		r, svc := newTestRouter(t)

		svc.EXPECT().
			GetUserByID(gomock.Any(), "bad-id").
			Return(nil, errors.New("not found"))

		req := httptest.NewRequest(http.MethodGet, "/users/bad-id", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandler_UpdateUser(t *testing.T) {
	t.Run("Should update user and return 200", func(t *testing.T) {
		r, svc := newTestRouter(t)

		expected := &dto.UserOutput{
			ID:        "user-1",
			FirstName: "Johnny",
			LastName:  "Doe",
			Email:     "johnny@example.com",
		}

		svc.EXPECT().
			UpdateUser(gomock.Any(), "user-1", "Johnny", "Doe", "johnny@example.com").
			Return(expected, nil)

		body, _ := json.Marshal(map[string]any{
			"first_name": "Johnny",
			"last_name":  "Doe",
			"email":      "johnny@example.com",
		})
		req := httptest.NewRequest(http.MethodPatch, "/users/user-1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var out dto.UserOutput
		require.NoError(t, json.NewDecoder(w.Body).Decode(&out))
		assert.Equal(t, expected.FirstName, out.FirstName)
	})

	t.Run("Should return 400 when body is invalid", func(t *testing.T) {
		r, _ := newTestRouter(t)

		req := httptest.NewRequest(http.MethodPatch, "/users/user-1", bytes.NewBufferString("not-json"))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Should return 500 when service returns error", func(t *testing.T) {
		r, svc := newTestRouter(t)

		svc.EXPECT().
			UpdateUser(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errors.New("db error"))

		body, _ := json.Marshal(map[string]any{"first_name": "X", "last_name": "Y", "email": "x@y.com"})
		req := httptest.NewRequest(http.MethodPatch, "/users/user-1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_DeleteUser(t *testing.T) {
	t.Run("Should delete user and return 204", func(t *testing.T) {
		r, svc := newTestRouter(t)

		svc.EXPECT().
			DeleteUser(gomock.Any(), "user-1").
			Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/users/user-1", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Should return 500 when service returns error", func(t *testing.T) {
		r, svc := newTestRouter(t)

		svc.EXPECT().
			DeleteUser(gomock.Any(), "bad-id").
			Return(errors.New("db error"))

		req := httptest.NewRequest(http.MethodDelete, "/users/bad-id", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
