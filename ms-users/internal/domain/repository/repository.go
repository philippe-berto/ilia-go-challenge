package repository

import (
	"context"
	"log/slog"
	"users/internal/dto"
	"users/internal/utils/postgres"

	"github.com/jmoiron/sqlx"
)

const (
	createUser     = "createUser"
	getUserByID    = "getUserByID"
	getUserByEmail = "getUserByEmail"
	getUsers       = "getUsers"
	updateUser     = "updateUser"
	deleteUser     = "deleteUser"
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

func (r *Repository) CreateUser(ctx context.Context, firstName, lastName, email, passwordHash string) (*dto.UserOutput, error) {
	if err := r.ensureStatements(); err != nil {
		return nil, err
	}

	var output dto.UserOutput
	err := r.statements[createUser].
		QueryRowContext(ctx, firstName, lastName, email, passwordHash).
		Scan(&output.ID, &output.FirstName, &output.LastName, &output.Email)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*dto.UserOutput, error) {
	if err := r.ensureStatements(); err != nil {
		return nil, err
	}

	var output dto.UserOutput
	err := r.statements[getUserByID].
		QueryRowContext(ctx, id).
		Scan(&output.ID, &output.FirstName, &output.LastName, &output.Email)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

type UserWithPassword struct {
	dto.UserOutput
	PasswordHash string
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*UserWithPassword, error) {
	if err := r.ensureStatements(); err != nil {
		return nil, err
	}

	var u UserWithPassword
	err := r.statements[getUserByEmail].
		QueryRowContext(ctx, email).
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.PasswordHash)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *Repository) GetUsers(ctx context.Context) ([]*dto.UserOutput, error) {
	if err := r.ensureStatements(); err != nil {
		return nil, err
	}

	rows, err := r.statements[getUsers].QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*dto.UserOutput
	for rows.Next() {
		var u dto.UserOutput
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Repository) UpdateUser(ctx context.Context, id, firstName, lastName, email string) (*dto.UserOutput, error) {
	if err := r.ensureStatements(); err != nil {
		return nil, err
	}

	var output dto.UserOutput
	err := r.statements[updateUser].
		QueryRowContext(ctx, firstName, lastName, email, id).
		Scan(&output.ID, &output.FirstName, &output.LastName, &output.Email)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func (r *Repository) DeleteUser(ctx context.Context, id string) error {
	if err := r.ensureStatements(); err != nil {
		return err
	}

	_, err := r.statements[deleteUser].ExecContext(ctx, id)
	return err
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
