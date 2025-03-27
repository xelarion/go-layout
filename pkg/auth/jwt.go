// Package auth provides authentication and authorization functionality.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/pkg/config"
)

// Common errors returned by JWT operations.
var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// JWTClaims represents the claims in a JWT token.
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWT represents a JWT authentication provider.
type JWT struct {
	config *config.JWT
	logger *zap.Logger
}

// NewJWT creates a new JWT provider.
func NewJWT(cfg *config.JWT, logger *zap.Logger) *JWT {
	return &JWT{
		config: cfg,
		logger: logger.Named("jwt"),
	}
}

// GenerateToken generates a new JWT token for a user.
func (j *JWT) GenerateToken(userID uint, role string) (string, time.Time, error) {
	now := time.Now()
	expiry := now.Add(j.config.Expiration)

	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.config.Secret))
	if err != nil {
		j.logger.Error("Failed to sign JWT token", zap.Error(err))
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, expiry, nil
}

// VerifyToken verifies a JWT token and returns the claims.
func (j *JWT) VerifyToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		j.logger.Warn("Failed to parse JWT token", zap.Error(err))
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken refreshes an existing token.
func (j *JWT) RefreshToken(tokenStr string) (string, time.Time, error) {
	claims, err := j.VerifyToken(tokenStr)
	if err != nil && !errors.Is(err, ErrExpiredToken) {
		return "", time.Time{}, err
	}

	// Generate a new token with the same claims but a new expiry
	return j.GenerateToken(claims.UserID, claims.Role)
}
