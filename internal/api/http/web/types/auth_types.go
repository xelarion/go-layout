// Package types contains request/response types for the API.
package types

import "time"

// LoginReq represents user login data.
type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	//Captcha   string `json:"captcha" binding:"required"`
	//CaptchaID string `json:"captcha_id" binding:"required"`
}

// LoginResp represents login response data.
type LoginResp struct {
	// JWT token for authorization
	Token string `json:"token"`
	// Time when the token will expire
	Expire time.Time `json:"expire"`
	// Time in seconds until the token expires (useful for frontend countdown)
	ExpiresIn int64 `json:"expires_in"`
	// Type of token (always "Bearer" for JWT)
	TokenType string `json:"token_type"`
}

// RefreshReq represents token refresh request.
type RefreshReq struct {
	// Expired or about-to-expire token to refresh
	Token string `json:"token" binding:"required"`
}

// RefreshResp represents token refresh response data.
type RefreshResp struct {
	// New JWT token for authorization
	Token string `json:"token"`
	// Time when the token will expire
	Expire time.Time `json:"expire"`
	// Time in seconds until the token expires
	ExpiresIn int64 `json:"expires_in"`
	// Type of token (always "Bearer" for JWT)
	TokenType string `json:"token_type"`
}

// CaptchaResp represents captcha data in responses.
type CaptchaResp struct {
	CaptchaID  string `json:"captcha_id"`
	CaptchaImg string `json:"captcha_img"` // Base64 encoded image
}
