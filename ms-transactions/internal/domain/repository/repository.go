package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"transactions/internal/domain/transaction"
	"transactions/internal/dto"
	"transactions/internal/utils/postgres"

	"github.com/jmoiron/sqlx"
)

const (
	createTransaction     = "createTransaction"
	getTransactionsByUser = "getTransactionsByUser"
	getTransactionsByType = "getTransactionsByType"
	getBalance            = "getBalance"
	setBalance            = "setBalance"
)

type Repository struct {
	db         *postgres.Client
	logger     *slog.Logger
	statements map[string]*sqlx.Stmt
}

func New(db *postgres.Client, logger *slog.Logger) (*Repository, error) {
	return &Repository{
		db:         db,
		logger:     logger,
		statements: map[string]*sqlx.Stmt{},
	}, nil
}

func (r *Repository) CreateTransaction(ctx context.Context, transaction *transaction.Transaction) (*dto.TransactionOutput, error) {
	if err := r.ensureStatements(); err != nil {
		return nil, err
	}

	tx, err := r.db.GetClient().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.logger.Error("Database: Error rolling back transaction", "error", rollbackErr)
			}
		}
	}()

	amount := transaction.Amount
	if transaction.Type == "debit" {
		amount = -amount
	}

	result, err := tx.StmtContext(ctx, r.statements[setBalance].Stmt).ExecContext(ctx, amount, transaction.UserID)
	if err != nil {
		return nil, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		err = fmt.Errorf("insufficient balance")
		return nil, err
	}

	var output dto.TransactionOutput
	err = tx.StmtContext(ctx, r.statements[createTransaction].Stmt).
		QueryRowContext(ctx, transaction.UserID, transaction.Type, transaction.Amount).
		Scan(&output.ID, &output.UserID, &output.Type, &output.Amount)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &output, nil
}

func (r *Repository) GetTransactionsByUser(ctx context.Context, userID string) ([]*dto.TransactionOutput, error) {
	if err := r.ensureStatements(); err != nil {
		return nil, err
	}

	rows, err := r.statements[getTransactionsByUser].QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*dto.TransactionOutput
	for rows.Next() {
		var t dto.TransactionOutput
		if err := rows.Scan(&t.ID, &t.UserID, &t.Type, &t.Amount); err != nil {
			return nil, err
		}
		transactions = append(transactions, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *Repository) GetTransactionsByType(ctx context.Context, userID, transactionType string) ([]*dto.TransactionOutput, error) {
	if err := r.ensureStatements(); err != nil {
		return nil, err
	}

	rows, err := r.statements[getTransactionsByType].QueryContext(ctx, userID, transactionType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*dto.TransactionOutput
	for rows.Next() {
		var t dto.TransactionOutput
		if err := rows.Scan(&t.ID, &t.UserID, &t.Type, &t.Amount); err != nil {
			return nil, err
		}
		transactions = append(transactions, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *Repository) GetBalance(ctx context.Context, userID string) (float64, error) {
	if err := r.ensureStatements(); err != nil {
		return 0, err
	}
	var balance float64
	err := r.statements[getBalance].QueryRowContext(ctx, userID).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (r *Repository) Close() error {
	for name, stmt := range r.statements {
		if err := stmt.Close(); err != nil {
			r.logger.Error("Database: Error closing statement", "statement", name, "error", err)
		}
	}
	return nil
}

func (r *Repository) prepareStatements() error {
	for _, item := range statementsList {
		stmt, err := r.db.PrepareStatement(item.query)
		if err != nil {
			r.logger.Error("Database: Error preparing statement", "statement", item.name, "error", err)
			return err
		}
		r.statements[item.name] = stmt
	}
	return nil
}

func (r *Repository) ensureStatements() error {
	if len(r.statements) == 0 {
		return r.prepareStatements()
	}
	return nil
}
