//go:build tool
// +build tool

package transactions

//go:generate mockgen -package mocks -destination mocks/service_mock.go transactions/internal/domain/http Service
//go:generate mockgen -package mocks -destination mocks/repository_mock.go transactions/internal/domain/service Repository
