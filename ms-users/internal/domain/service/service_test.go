package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"users/internal/domain/repository"
	"users/internal/dto"
	"users/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func newTestService(t *testing.T) (*Service, *mocks.MockRepository, *mocks.MockTokenGenerator) {
	t.Helper()
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	jwt := mocks.NewMockTokenGenerator(ctrl)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return New(repo, jwt, logger), repo, jwt
}

func TestService_CreateUser(t *testing.T) {
	t.Run("Should create user and return output", func(t *testing.T) {
		svc, repo, _ := newTestService(t)

		expected := &dto.UserOutput{
			ID:        "user-1",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		}

		repo.EXPECT().
			CreateUser(gomock.Any(), "John", "Doe", "john@example.com", gomock.Any()).
			Return(expected, nil)

		out, err := svc.CreateUser(context.Background(), "John", "Doe", "john@example.com", "secret123")
		require.NoError(t, err)
		assert.Equal(t, expected, out)
	})

	t.Run("Should return error when repository fails", func(t *testing.T) {
		svc, repo, _ := newTestService(t)

		repo.EXPECT().
			CreateUser(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errors.New("email already exists"))

		_, err := svc.CreateUser(context.Background(), "John", "Doe", "john@example.com", "secret123")
		require.EqualError(t, err, "email already exists")
	})
}

func TestService_GetUser(t *testing.T) {
	t.Run("Should return user by ID", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		userID := "user-1"

		expected := &dto.UserOutput{
			ID:        userID,
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		}

		repo.EXPECT().
			GetUserByID(gomock.Any(), userID).
			Return(expected, nil)

		out, err := svc.GetUserByID(context.Background(), userID)
		require.NoError(t, err)
		assert.Equal(t, expected, out)
	})

	t.Run("Should return error when user is not found", func(t *testing.T) {
		svc, repo, _ := newTestService(t)

		repo.EXPECT().
			GetUserByID(gomock.Any(), "unknown-id").
			Return(nil, errors.New("sql: no rows in result set"))

		_, err := svc.GetUserByID(context.Background(), "unknown-id")
		require.Error(t, err)
	})

	t.Run("Should return all users", func(t *testing.T) {
		svc, repo, _ := newTestService(t)

		expected := []*dto.UserOutput{
			{ID: "user-1", FirstName: "John", LastName: "Doe", Email: "john@example.com"},
			{ID: "user-2", FirstName: "Jane", LastName: "Doe", Email: "jane@example.com"},
		}

		repo.EXPECT().
			GetUsers(gomock.Any()).
			Return(expected, nil)

		result, err := svc.GetUsers(context.Background())
		require.NoError(t, err)
		assert.Len(t, result, len(expected))
	})
}

func TestService_UpdateUser(t *testing.T) {
	t.Run("Should update and return the updated user", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		userID := "user-1"

		expected := &dto.UserOutput{
			ID:        userID,
			FirstName: "Johnny",
			LastName:  "Doe",
			Email:     "johnny@example.com",
		}

		repo.EXPECT().
			UpdateUser(gomock.Any(), userID, "Johnny", "Doe", "johnny@example.com").
			Return(expected, nil)

		out, err := svc.UpdateUser(context.Background(), userID, "Johnny", "Doe", "johnny@example.com")
		require.NoError(t, err)
		assert.Equal(t, expected, out)
	})

	t.Run("Should return error when update fails", func(t *testing.T) {
		svc, repo, _ := newTestService(t)

		repo.EXPECT().
			UpdateUser(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errors.New("user not found"))

		_, err := svc.UpdateUser(context.Background(), "bad-id", "X", "Y", "x@y.com")
		require.EqualError(t, err, "user not found")
	})
}

func TestService_DeleteUser(t *testing.T) {
	t.Run("Should delete user successfully", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		userID := "user-1"

		repo.EXPECT().
			DeleteUser(gomock.Any(), userID).
			Return(nil)

		err := svc.DeleteUser(context.Background(), userID)
		require.NoError(t, err)
	})

	t.Run("Should return error when delete fails", func(t *testing.T) {
		svc, repo, _ := newTestService(t)

		repo.EXPECT().
			DeleteUser(gomock.Any(), "bad-id").
			Return(errors.New("user not found"))

		err := svc.DeleteUser(context.Background(), "bad-id")
		require.EqualError(t, err, "user not found")
	})
}

func TestService_Authenticate(t *testing.T) {
	t.Run("Should return token on valid credentials", func(t *testing.T) {
		svc, repo, jwt := newTestService(t)

		hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
		require.NoError(t, err)

		repo.EXPECT().
			GetUserByEmail(gomock.Any(), "john@example.com").
			Return(&repository.UserWithPassword{
				UserOutput:   dto.UserOutput{ID: "user-1"},
				PasswordHash: string(hash),
			}, nil)

		jwt.EXPECT().
			GenerateToken("user-1").
			Return("token-abc", nil)

		out, err := svc.Authenticate(context.Background(), "john@example.com", "secret123")
		require.NoError(t, err)
		assert.Equal(t, "token-abc", out.Token)
	})

	t.Run("Should return error when user is not found", func(t *testing.T) {
		svc, repo, _ := newTestService(t)

		repo.EXPECT().
			GetUserByEmail(gomock.Any(), "nobody@example.com").
			Return(nil, errors.New("not found"))

		_, err := svc.Authenticate(context.Background(), "nobody@example.com", "pass")
		require.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("Should return error when password is wrong", func(t *testing.T) {
		svc, repo, _ := newTestService(t)

		hash, err := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
		require.NoError(t, err)

		repo.EXPECT().
			GetUserByEmail(gomock.Any(), "john@example.com").
			Return(&repository.UserWithPassword{
				UserOutput:   dto.UserOutput{ID: "user-1"},
				PasswordHash: string(hash),
			}, nil)

		_, err = svc.Authenticate(context.Background(), "john@example.com", "wrong")
		require.ErrorIs(t, err, ErrInvalidCredentials)
	})
}
