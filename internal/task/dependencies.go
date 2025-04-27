// Package task provides common functionality for all task types.
package task

import (
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/repository"
	"github.com/xelarion/go-layout/internal/usecase"
)

// Dependencies contains all dependencies needed by tasks.
type Dependencies struct {
	// Core infrastructure (unexported)
	data   *repository.Data
	logger *zap.Logger

	// Repositories
	UserRepo usecase.UserRepository

	// Usecases
	UserUseCase *usecase.UserUseCase
}

// NewDependencies creates a new dependencies instance with all required dependencies.
func NewDependencies(data *repository.Data, logger *zap.Logger) (*Dependencies, error) {
	// Create repositories
	departmentRepo := repository.NewDepartmentRepository(data)
	roleRepo := repository.NewRoleRepository(data)
	userRepo := repository.NewUserRepository(data)

	// Create usecases
	userUseCase := usecase.NewUserUseCase(data, userRepo, roleRepo, departmentRepo)

	deps := &Dependencies{
		// Core infrastructure
		data:   data,
		logger: logger,

		// Repositories
		UserRepo: userRepo,

		// Usecases
		UserUseCase: userUseCase,
	}

	return deps, nil
}

// Data returns the data access layer
func (d *Dependencies) Data() *repository.Data {
	return d.data
}

// Logger returns the logger
func (d *Dependencies) Logger() *zap.Logger {
	return d.logger
}
