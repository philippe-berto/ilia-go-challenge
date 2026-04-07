//go:build integration
// +build integration

package repository

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"users/internal/utils/config"
	"users/internal/utils/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRepo(t *testing.T) *Repository {
	t.Helper()

	cfg := config.PostgresConfig{
		Host:         getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Name:         getEnvOrDefault("POSTGRES_DB", "users"),
		Password:     getEnvOrDefault("POSTGRES_PASSWORD", "postgres"),
		User:         getEnvOrDefault("POSTGRES_USER", "postgres"),
		Port:         5432,
		Driver:       "postgres",
		Timeout:      5,
		RunMigration: false,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	db, err := postgres.New(context.Background(), cfg, "")
	require.NoError(t, err, "failed to connect to database")
	t.Cleanup(func() { db.Close() })

	repo, err := New(db, logger)
	require.NoError(t, err, "failed to create repository")

	return repo
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func cleanupUser(t *testing.T, repo *Repository, userID string) {
	t.Helper()
	db := repo.db.GetClient()
	if _, err := db.Exec("DELETE FROM users WHERE id = $1", userID); err != nil {
		t.Logf("cleanup: failed to delete user %s: %v", userID, err)
	}
}

func TestIntegration_CreateUser(t *testing.T) {
	t.Run("Should create a user and return the output", func(t *testing.T) {
		repo := newTestRepo(t)
		email := uuid.New().String() + "@test.com"

		out, err := repo.CreateUser(context.Background(), "John", "Doe", email, "hashed_pw")
		require.NoError(t, err)
		t.Cleanup(func() { cleanupUser(t, repo, out.ID) })

		assert.NotEmpty(t, out.ID)
		assert.Equal(t, "John", out.FirstName)
		assert.Equal(t, "Doe", out.LastName)
		assert.Equal(t, email, out.Email)
	})

	t.Run("Should fail when email already exists", func(t *testing.T) {
		repo := newTestRepo(t)
		email := uuid.New().String() + "@test.com"

		out, err := repo.CreateUser(context.Background(), "Jane", "Doe", email, "hashed_pw")
		require.NoError(t, err, "setup: first insert failed")
		t.Cleanup(func() { cleanupUser(t, repo, out.ID) })

		_, err = repo.CreateUser(context.Background(), "Jane", "Doe", email, "hashed_pw")
		require.Error(t, err)
	})
}

func TestIntegration_GetUser(t *testing.T) {
	t.Run("Should get a user by ID", func(t *testing.T) {
		repo := newTestRepo(t)
		email := uuid.New().String() + "@test.com"

		created, err := repo.CreateUser(context.Background(), "Alice", "Smith", email, "hashed_pw")
		require.NoError(t, err)
		t.Cleanup(func() { cleanupUser(t, repo, created.ID) })

		out, err := repo.GetUserByID(context.Background(), created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, out.ID)
		assert.Equal(t, "Alice", out.FirstName)
		assert.Equal(t, email, out.Email)
	})

	t.Run("Should return error when user ID does not exist", func(t *testing.T) {
		repo := newTestRepo(t)

		_, err := repo.GetUserByID(context.Background(), uuid.New().String())
		require.Error(t, err)
	})

	t.Run("Should get a user by email", func(t *testing.T) {
		repo := newTestRepo(t)
		email := uuid.New().String() + "@test.com"

		created, err := repo.CreateUser(context.Background(), "Bob", "Jones", email, "secret_hash")
		require.NoError(t, err)
		t.Cleanup(func() { cleanupUser(t, repo, created.ID) })

		out, err := repo.GetUserByEmail(context.Background(), email)
		require.NoError(t, err)
		assert.Equal(t, created.ID, out.ID)
		assert.Equal(t, "secret_hash", out.PasswordHash)
	})

	t.Run("Should return error when email does not exist", func(t *testing.T) {
		repo := newTestRepo(t)

		_, err := repo.GetUserByEmail(context.Background(), uuid.New().String()+"@missing.com")
		require.Error(t, err)
	})
}

func TestIntegration_GetUsers(t *testing.T) {
	t.Run("Should return all users including newly created ones", func(t *testing.T) {
		repo := newTestRepo(t)

		ids := make([]string, 3)
		for i := range ids {
			email := uuid.New().String() + "@test.com"
			out, err := repo.CreateUser(context.Background(), "User", "Test", email, "pw")
			require.NoError(t, err)
			ids[i] = out.ID
		}
		t.Cleanup(func() {
			for _, id := range ids {
				cleanupUser(t, repo, id)
			}
		})

		users, err := repo.GetUsers(context.Background())
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 3)
	})
}

func TestIntegration_UpdateUser(t *testing.T) {
	t.Run("Should update user fields and return updated output", func(t *testing.T) {
		repo := newTestRepo(t)
		email := uuid.New().String() + "@test.com"
		newEmail := uuid.New().String() + "@updated.com"

		created, err := repo.CreateUser(context.Background(), "Old", "Name", email, "pw")
		require.NoError(t, err)
		t.Cleanup(func() { cleanupUser(t, repo, created.ID) })

		out, err := repo.UpdateUser(context.Background(), created.ID, "New", "Name", newEmail)
		require.NoError(t, err)
		assert.Equal(t, created.ID, out.ID)
		assert.Equal(t, "New", out.FirstName)
		assert.Equal(t, newEmail, out.Email)
	})

	t.Run("Should return error when user ID does not exist", func(t *testing.T) {
		repo := newTestRepo(t)

		_, err := repo.UpdateUser(context.Background(), uuid.New().String(), "X", "Y", uuid.New().String()+"@z.com")
		require.Error(t, err)
	})
}

func TestIntegration_DeleteUser(t *testing.T) {
	t.Run("Should delete a user successfully", func(t *testing.T) {
		repo := newTestRepo(t)
		email := uuid.New().String() + "@test.com"

		created, err := repo.CreateUser(context.Background(), "Del", "User", email, "pw")
		require.NoError(t, err)

		err = repo.DeleteUser(context.Background(), created.ID)
		require.NoError(t, err)

		_, err = repo.GetUserByID(context.Background(), created.ID)
		require.Error(t, err, "user should not be found after deletion")
	})

	t.Run("Should not error when deleting a non-existent user", func(t *testing.T) {
		repo := newTestRepo(t)

		err := repo.DeleteUser(context.Background(), uuid.New().String())
		require.NoError(t, err)
	})
}
