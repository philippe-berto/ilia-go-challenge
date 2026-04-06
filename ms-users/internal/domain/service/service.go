package service

import (
	"context"
	"errors"
	"log/slog"
	"users/internal/domain/repository"
	"users/internal/dto"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type TokenGenerator interface {
	GenerateToken(userID string) (string, error)
}

type Repository interface {
	CreateUser(ctx context.Context, firstName, lastName, email, passwordHash string) (*dto.UserOutput, error)
	GetUserByID(ctx context.Context, id string) (*dto.UserOutput, error)
	GetUserByEmail(ctx context.Context, email string) (*repository.UserWithPassword, error)
	GetUsers(ctx context.Context) ([]*dto.UserOutput, error)
	UpdateUser(ctx context.Context, id, firstName, lastName, email string) (*dto.UserOutput, error)
	DeleteUser(ctx context.Context, id string) error
}

type Service struct {
	repo   Repository
	jwt    TokenGenerator
	logger *slog.Logger
}

func New(repo Repository, jwt TokenGenerator, logger *slog.Logger) *Service {
	return &Service{repo: repo, jwt: jwt, logger: logger}
}

func (s *Service) CreateUser(ctx context.Context, firstName, lastName, email, password string) (*dto.UserOutput, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return nil, err
	}

	output, err := s.repo.CreateUser(ctx, firstName, lastName, email, string(hash))
	if err != nil {
		s.logger.Error("failed to create user", "error", err)
		return nil, err
	}

	return output, nil
}

func (s *Service) GetUserByID(ctx context.Context, id string) (*dto.UserOutput, error) {
	output, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get user by id", "error", err)
		return nil, err
	}

	return output, nil
}

func (s *Service) GetUsers(ctx context.Context) ([]*dto.UserOutput, error) {
	users, err := s.repo.GetUsers(ctx)
	if err != nil {
		s.logger.Error("failed to get users", "error", err)
		return nil, err
	}

	return users, nil
}

func (s *Service) UpdateUser(ctx context.Context, id, firstName, lastName, email string) (*dto.UserOutput, error) {
	output, err := s.repo.UpdateUser(ctx, id, firstName, lastName, email)
	if err != nil {
		s.logger.Error("failed to update user", "error", err)
		return nil, err
	}

	return output, nil
}

func (s *Service) DeleteUser(ctx context.Context, id string) error {
	if err := s.repo.DeleteUser(ctx, id); err != nil {
		s.logger.Error("failed to delete user", "error", err)
		return err
	}

	return nil
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (*dto.AuthOutput, error) {
	u, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.jwt.GenerateToken(u.ID)
	if err != nil {
		s.logger.Error("failed to generate token", "error", err)
		return nil, err
	}

	return &dto.AuthOutput{Token: token}, nil
}
