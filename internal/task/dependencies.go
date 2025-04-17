// Package task provides common functionality for all task types.
package task

import (
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/repository"
	"github.com/xelarion/go-layout/internal/usecase"
)

// Dependencies contains all dependencies needed by tasks.
type Dependencies struct {
	// Core infrastructure (unexported)
	db          *gorm.DB
	redisClient *redis.Client
	logger      *zap.Logger

	// Repositories
	UserRepo usecase.UserRepository

	// Usecases
	UserUseCase *usecase.UserUseCase
}

// NewDependencies creates a new dependencies instance with all required dependencies.
func NewDependencies(db *gorm.DB, redisClient *redis.Client, logger *zap.Logger) *Dependencies {
	// Create repositories
	data := repository.NewData(db, redisClient)
	userRepo := repository.NewUserRepository(data)
	roleRepo := repository.NewRoleRepository(data)
	departmentRepo := repository.NewDepartmentRepository(data)

	// Create usecases
	userUseCase := usecase.NewUserUseCase(data, userRepo, roleRepo, departmentRepo)

	return &Dependencies{
		// Core infrastructure
		db:          db,
		redisClient: redisClient,
		logger:      logger,

		// Repositories
		UserRepo: userRepo,

		// Usecases
		UserUseCase: userUseCase,
	}
}

// DB returns the database connection
func (d *Dependencies) DB() *gorm.DB {
	return d.db
}

// RedisClient returns the Redis client
func (d *Dependencies) RedisClient() *redis.Client {
	return d.redisClient
}

// Logger returns the logger
func (d *Dependencies) Logger() *zap.Logger {
	return d.logger
}
