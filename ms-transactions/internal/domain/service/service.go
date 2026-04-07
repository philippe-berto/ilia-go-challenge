package service

import (
	"context"
	"log/slog"
	"transactions/internal/domain/transaction"
	"transactions/internal/dto"
)

type Repository interface {
	CreateTransaction(ctx context.Context, transaction *transaction.Transaction) (*dto.TransactionOutput, error)
	GetTransactionsByUser(ctx context.Context, userID string) ([]*dto.TransactionOutput, error)
	GetTransactionsByType(ctx context.Context, userID, transactionType string) ([]*dto.TransactionOutput, error)
	GetBalance(ctx context.Context, userID string) (float64, error)
}

type Service struct {
	repo   Repository
	logger *slog.Logger
}

func New(repo Repository, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

func (s *Service) CreateTransaction(ctx context.Context, transaction *transaction.Transaction) (*dto.TransactionOutput, error) {
	output, err := s.repo.CreateTransaction(ctx, transaction)
	if err != nil {
		s.logger.Error("failed to create transaction", "error", err)
		return nil, err
	}

	return output, nil
}

func (s *Service) GetTransactionsByUser(ctx context.Context, userID string) ([]*dto.TransactionOutput, error) {
	transactions, err := s.repo.GetTransactionsByUser(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get transactions by user", "error", err)
		return nil, err
	}

	return transactions, nil
}

func (s *Service) GetTransactionsByType(ctx context.Context, userID, transactionType string) ([]*dto.TransactionOutput, error) {
	transactions, err := s.repo.GetTransactionsByType(ctx, userID, transactionType)
	if err != nil {
		s.logger.Error("failed to get transactions by type", "error", err)
		return nil, err
	}

	return transactions, nil
}

func (s *Service) GetBalance(ctx context.Context, userID string) (float64, error) {
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get balance", "error", err)
		return 0, err
	}

	return balance, nil
}
