package types

type NewCaptchaReq struct {
}

type NewCaptchaResp struct {
	ID    string `json:"id"`
	Image string `json:"image"` // base64
}

type ReloadCaptchaReq struct {
	ID string `uri:"id" binding:"required" swaggerignore:"true"`
}

type ReloadCaptchaResp struct {
	Image string `json:"image"` // base64
}

type GetRSAPublicKeyReq struct {
}

type GetRSAPublicKeyResp struct {
	PubKey string `json:"pub_key"` // public key
	Key    string `json:"key"`     // key
}

// LoginReq represents user login data.
type LoginReq struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Captcha   string `json:"captcha" binding:"required"`
	CaptchaID string `json:"captcha_id" binding:"required"`
	Key       string `json:"key" binding:"required"` // rsa redis key
}

// LoginResp represents login response data.
type LoginResp struct {
	// JWT token for authorization
	Token string `json:"token"`
	// Time when the token will expire
	Expire Time `json:"expire"`
	// Time in seconds until the token expires (useful for frontend countdown)
	ExpiresIn int64 `json:"expires_in"`
	// Type of token (always "Bearer" for JWT)
	TokenType string `json:"token_type"`
}

// RefreshTokenReq represents token refresh request.
type RefreshTokenReq struct {
	// Expired or about-to-expire token to refresh
	Token string `json:"token" binding:"required"`
}

// RefreshTokenResp represents token refresh response data.
type RefreshTokenResp struct {
	// New JWT token for authorization
	Token string `json:"token"`
	// Time when the token will expire
	Expire Time `json:"expire"`
	// Time in seconds until the token expires
	ExpiresIn int64 `json:"expires_in"`
	// Type of token (always "Bearer" for JWT)
	TokenType string `json:"token_type"`
}

// LogoutReq represents logout request.
type LogoutReq struct {
}

// LogoutResp represents logout response data.
type LogoutResp struct {
}

type GetProfileReq struct {
}

type GetProfileResp struct {
	ID        uint   `json:"id"`
	CreatedAt Time   `json:"created_at"`
	Username  string `json:"username"`
	Email     string `json:"email"`
}

// UpdateProfileReq represents profile update data.
type UpdateProfileReq struct {
	Username string `json:"username" binding:"omitempty,min=2,max=100"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=6"`
	Avatar   string `json:"avatar" binding:"omitempty,url"`
}

// UpdateProfileResp represents profile update response.
type UpdateProfileResp struct {
}

type GetCurrentUserInfoReq struct {
}

type GetCurrentUserInfoResp struct {
	ID          uint     `json:"id"`
	RoleSlug    string   `json:"role_slug"`
	Permissions []string `json:"permissions"`
}
