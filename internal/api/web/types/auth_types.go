// Package types contains request and response types for the web API.
package types

import "time"

// LoginReq represents user login data.
type LoginReq struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
	Captcha   string `json:"captcha" binding:"required"`
	CaptchaID string `json:"captcha_id" binding:"required"`
}

// LoginResp represents login response data.
type LoginResp struct {
	Token       string    `json:"token"`
	TokenExpiry time.Time `json:"token_expiry"`
}

// CaptchaResp represents captcha data in responses.
type CaptchaResp struct {
	CaptchaID  string `json:"captcha_id"`
	CaptchaImg string `json:"captcha_img"` // Base64 encoded image
}
