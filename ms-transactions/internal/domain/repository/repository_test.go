//go:build integration
// +build integration

package repository

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"transactions/internal/domain/transaction"
	"transactions/internal/utils/config"
	"transactions/internal/utils/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRepo(t *testing.T) *Repository {
	t.Helper()

	cfg := config.PostgresConfig{
		Host:         getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Name:         getEnvOrDefault("POSTGRES_DB", "transactions"),
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
	if _, err := db.Exec("DELETE FROM transactions WHERE user_id = $1", userID); err != nil {
		t.Logf("cleanup: failed to delete transactions for user %s: %v", userID, err)
	}
	if _, err := db.Exec("DELETE FROM accounts WHERE user_id = $1", userID); err != nil {
		t.Logf("cleanup: failed to delete account for user %s: %v", userID, err)
	}
}

func TestIntegration_CreateTransaction(t *testing.T) {
	t.Run("Should create a credit transaction successfully", func(t *testing.T) {
		repo := newTestRepo(t)
		userID := uuid.New().String()
		t.Cleanup(func() { cleanupUser(t, repo, userID) })

		out, err := repo.CreateTransaction(context.Background(), &transaction.Transaction{
			UserID: userID,
			Type:   transaction.Credit,
			Amount: 100.00,
		})
		require.NoError(t, err)
		assert.Equal(t, userID, out.UserID)
		assert.Equal(t, 100.00, out.Amount)
		assert.Equal(t, string(transaction.Credit), out.Type)
	})

	t.Run("Should fail a debit transaction when balance is insufficient", func(t *testing.T) {
		repo := newTestRepo(t)
		userID := uuid.New().String()
		t.Cleanup(func() { cleanupUser(t, repo, userID) })

		_, err := repo.CreateTransaction(context.Background(), &transaction.Transaction{
			UserID: userID,
			Type:   transaction.Credit,
			Amount: 50.00,
		})
		require.NoError(t, err, "setup credit failed")

		_, err = repo.CreateTransaction(context.Background(), &transaction.Transaction{
			UserID: userID,
			Type:   transaction.Debit,
			Amount: 200.00,
		})
		require.EqualError(t, err, "insufficient balance")
	})
}

func TestIntegration_GetTransactions(t *testing.T) {
	t.Run("Should return all transactions for a user", func(t *testing.T) {
		repo := newTestRepo(t)
		userID := uuid.New().String()
		t.Cleanup(func() { cleanupUser(t, repo, userID) })

		amounts := []float64{100.00, 200.00, 50.00}
		for _, a := range amounts {
			_, err := repo.CreateTransaction(context.Background(), &transaction.Transaction{
				UserID: userID,
				Type:   transaction.Credit,
				Amount: a,
			})
			require.NoError(t, err, "failed to create transaction")
		}

		txs, err := repo.GetTransactionsByUser(context.Background(), userID)
		require.NoError(t, err)
		assert.Len(t, txs, len(amounts))
	})

	t.Run("Should return only credit transactions when filtering by type", func(t *testing.T) {
		repo := newTestRepo(t)
		userID := uuid.New().String()
		t.Cleanup(func() { cleanupUser(t, repo, userID) })

		for _, a := range []float64{300.00, 200.00} {
			_, err := repo.CreateTransaction(context.Background(), &transaction.Transaction{
				UserID: userID,
				Type:   transaction.Credit,
				Amount: a,
			})
			require.NoError(t, err, "failed to create credit")
		}

		credits, err := repo.GetTransactionsByType(context.Background(), userID, string(transaction.Credit))
		require.NoError(t, err)
		assert.Len(t, credits, 2)
	})

	t.Run("Should return only debit transactions when filtering by type", func(t *testing.T) {
		repo := newTestRepo(t)
		userID := uuid.New().String()
		t.Cleanup(func() { cleanupUser(t, repo, userID) })

		_, err := repo.CreateTransaction(context.Background(), &transaction.Transaction{
			UserID: userID,
			Type:   transaction.Credit,
			Amount: 300.00,
		})
		require.NoError(t, err, "failed to create credit")

		_, err = repo.CreateTransaction(context.Background(), &transaction.Transaction{
			UserID: userID,
			Type:   transaction.Debit,
			Amount: 100.00,
		})
		require.NoError(t, err, "failed to create debit")

		debits, err := repo.GetTransactionsByType(context.Background(), userID, string(transaction.Debit))
		require.NoError(t, err)
		assert.Len(t, debits, 1)
	})
}

func TestIntegration_GetBalance(t *testing.T) {
	t.Run("Should return the correct balance after credits and debits", func(t *testing.T) {
		repo := newTestRepo(t)
		userID := uuid.New().String()
		t.Cleanup(func() { cleanupUser(t, repo, userID) })

		_, err := repo.CreateTransaction(context.Background(), &transaction.Transaction{
			UserID: userID,
			Type:   transaction.Credit,
			Amount: 200.00,
		})
		require.NoError(t, err, "failed to create credit")

		_, err = repo.CreateTransaction(context.Background(), &transaction.Transaction{
			UserID: userID,
			Type:   transaction.Debit,
			Amount: 50.00,
		})
		require.NoError(t, err, "failed to create debit")

		balance, err := repo.GetBalance(context.Background(), userID)
		require.NoError(t, err)
		assert.Equal(t, 150.00, balance)
	})
}
