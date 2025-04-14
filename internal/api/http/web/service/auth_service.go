package service

import (
	"bytes"
	"context"
	"encoding/base64"

	"github.com/dchest/captcha"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/middleware"
	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/internal/usecase"
	"github.com/xelarion/go-layout/pkg/errs"
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

func (s *AuthService) GetRSAPublicKey(ctx context.Context, req *types.GetRSAPublicKeyReq) (*types.GetRSAPublicKeyResp, error) {
	publicKey, rdsKey, err := s.userUseCase.GetRSAPublicKey(ctx)
	if err != nil {
		return nil, err
	}

	return &types.GetRSAPublicKeyResp{
		PubKey: publicKey,
		Key:    rdsKey,
	}, nil
}

func (s *AuthService) NewCaptcha(ctx context.Context, req *types.NewCaptchaReq) (*types.NewCaptchaResp, error) {
	id := captcha.NewLen(4)
	var buf bytes.Buffer
	if err := captcha.WriteImage(&buf, id, 180, 80); err != nil {
		return nil, errs.WrapInternal(err, "failed to write captcha image")
	}

	return &types.NewCaptchaResp{
		ID:    id,
		Image: "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}

func (s *AuthService) ReloadCaptcha(ctx context.Context, req *types.ReloadCaptchaReq) (*types.ReloadCaptchaResp, error) {
	if !captcha.Reload(req.ID) {
		return nil, errs.NewBusiness("failed to reload captcha")
	}

	var buf bytes.Buffer
	if err := captcha.WriteImage(&buf, req.ID, 180, 80); err != nil {
		return nil, errs.WrapInternal(err, "failed to write captcha image")
	}

	return &types.ReloadCaptchaResp{
		Image: "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}

func (s *AuthService) GetCurrentUserInfo(ctx context.Context, req *types.GetCurrentUserInfoReq) (*types.GetCurrentUserInfoResp, error) {
	current := middleware.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("invalid credentials").WithReason(errs.ReasonUnauthorized)
	}

	user, err := s.userUseCase.GetByID(ctx, current.UserID)
	if err != nil {
		return nil, err
	}

	return &types.GetCurrentUserInfoResp{
		ID:       user.ID,
		RoleSlug: current.RoleSlug,
		Permissions: []string{
			"page:home",
			"menu:setting",
			"page:roles",
			"page:users",
			"page:departments",
		}, // TODO
	}, nil
}

// GetProfile gets the current user's profile.
func (s *AuthService) GetProfile(ctx context.Context, req *types.GetProfileReq) (*types.GetProfileResp, error) {
	current := middleware.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("invalid credentials").WithReason(errs.ReasonUnauthorized)
	}

	user, err := s.userUseCase.GetByID(ctx, current.UserID)
	if err != nil {
		return nil, err
	}

	return &types.GetProfileResp{
		ID:        user.ID,
		CreatedAt: types.Time(user.CreatedAt),
		Username:  user.Username,
		Email:     user.Email,
	}, nil
}

// UpdateProfile updates the current user's profile.
func (s *AuthService) UpdateProfile(ctx context.Context, req *types.UpdateProfileReq) (*types.UpdateProfileResp, error) {
	current := middleware.GetCurrent(ctx)
	if current == nil {
		return nil, errs.NewBusiness("invalid credentials").WithReason(errs.ReasonUnauthorized)
	}

	// First check if the user exists
	_, err := s.userUseCase.GetByID(ctx, current.UserID)
	if err != nil {
		return nil, err
	}

	// Create update params
	params := usecase.UpdateUserParams{
		ID: current.UserID,
	}

	// Only set fields that are provided in the request
	if req.Username != "" {
		params.Username = req.Username
		params.UsernameSet = true
	}

	if req.Email != "" {
		params.Email = req.Email
		params.EmailSet = true
	}

	if req.Password != "" {
		params.Password = req.Password
		params.PasswordSet = true
	}

	if err := s.userUseCase.Update(ctx, params); err != nil {
		return nil, err
	}

	return &types.UpdateProfileResp{}, nil
}
