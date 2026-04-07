package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"transactions/internal/domain/repository"
	"transactions/internal/domain/transaction"
	"transactions/internal/dto"
	"transactions/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestService(t *testing.T) (*Service, *mocks.MockRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return New(repo, logger), repo
}

func TestService_CreateTransaction(t *testing.T) {
	t.Run("Should create a credit transaction successfully", func(t *testing.T) {
		svc, repo := newTestService(t)
		userID := "user-123"
		txID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"

		expected := &dto.TransactionOutput{
			ID:     txID,
			UserID: userID,
			Type:   "CREDIT",
			Amount: 100.00,
		}

		repo.EXPECT().
			CreateTransaction(gomock.Any(), &transaction.Transaction{
				ID:     txID,
				UserID: userID,
				Type:   transaction.Credit,
				Amount: 100.00,
			}).
			Return(expected, nil)

		out, err := svc.CreateTransaction(context.Background(), &transaction.Transaction{
			ID:     txID,
			UserID: userID,
			Type:   transaction.Credit,
			Amount: 100.00,
		})
		require.NoError(t, err)
		assert.Equal(t, expected, out)
	})

	t.Run("Should return error when balance is insufficient", func(t *testing.T) {
		svc, repo := newTestService(t)
		userID := "user-123"
		txID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"

		repo.EXPECT().
			CreateTransaction(gomock.Any(), &transaction.Transaction{
				ID:     txID,
				UserID: userID,
				Type:   transaction.Debit,
				Amount: 200.00,
			}).
			Return(nil, repository.ErrInsufficientBalance)

		_, err := svc.CreateTransaction(context.Background(), &transaction.Transaction{
			ID:     txID,
			UserID: userID,
			Type:   transaction.Debit,
			Amount: 200.00,
		})
		require.ErrorIs(t, err, repository.ErrInsufficientBalance)
	})
}

func TestService_GetTransactions(t *testing.T) {
	t.Run("Should return all transactions for a user", func(t *testing.T) {
		svc, repo := newTestService(t)
		userID := "user-123"

		expected := []*dto.TransactionOutput{
			{ID: "tx-1", UserID: userID, Type: "CREDIT", Amount: 100.00},
			{ID: "tx-2", UserID: userID, Type: "CREDIT", Amount: 200.00},
		}

		repo.EXPECT().
			GetTransactionsByUser(gomock.Any(), userID).
			Return(expected, nil)

		result, err := svc.GetTransactionsByUser(context.Background(), userID)
		require.NoError(t, err)
		assert.Len(t, result, len(expected))
	})

	t.Run("Should return transactions filtered by type", func(t *testing.T) {
		svc, repo := newTestService(t)
		userID := "user-123"

		expected := []*dto.TransactionOutput{
			{ID: "tx-1", UserID: userID, Type: "CREDIT", Amount: 100.00},
		}

		repo.EXPECT().
			GetTransactionsByType(gomock.Any(), userID, "CREDIT").
			Return(expected, nil)

		result, err := svc.GetTransactionsByType(context.Background(), userID, "CREDIT")
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})
}

func TestService_GetBalance(t *testing.T) {
	t.Run("Should return the correct balance", func(t *testing.T) {
		svc, repo := newTestService(t)
		userID := "user-123"

		repo.EXPECT().
			GetBalance(gomock.Any(), userID).
			Return(150.00, nil)

		balance, err := svc.GetBalance(context.Background(), userID)
		require.NoError(t, err)
		assert.Equal(t, 150.00, balance)
	})
}
