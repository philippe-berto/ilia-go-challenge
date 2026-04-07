//go:build tool
// +build tool

package users

//go:generate mockgen -package mocks -destination mocks/service_mock.go users/internal/domain/http Service
//go:generate mockgen -package mocks -destination mocks/repository_mock.go users/internal/domain/service Repository
//go:generate mockgen -package mocks -destination mocks/token_generator_mock.go users/internal/domain/service TokenGenerator
