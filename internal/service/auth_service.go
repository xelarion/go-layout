package service

import (
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/usecase"
)

type AuthService struct {
	userUseCase *usecase.UserUseCase
	logger      *zap.Logger
}

// NewAuthService creates a new instance of AuthService.
func NewAuthService(userUseCase *usecase.UserUseCase, logger *zap.Logger) *AuthService {
	return &AuthService{
		userUseCase: userUseCase,
		logger:      logger.Named("auth_service"),
	}
}
