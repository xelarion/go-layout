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

func (u *AuthService) GetRSAPublicKey(ctx context.Context, req *types.GetRSAPublicKeyReq) (*types.GetRSAPublicKeyResp, error) {
	publicKey, rdsKey, err := u.userUseCase.GetRSAPublicKey(ctx)
	if err != nil {
		return nil, err
	}

	return &types.GetRSAPublicKeyResp{
		PubKey: publicKey,
		Key:    rdsKey,
	}, nil
}

func (u *AuthService) NewCaptcha(ctx context.Context, req *types.NewCaptcha) (*types.NewCaptchaResp, error) {
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

func (u *AuthService) ReloadCaptcha(ctx context.Context, req *types.ReloadCaptchaReq) (*types.ReloadCaptchaResp, error) {
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

func (u *AuthService) GetCurrentUserInfo(ctx context.Context, req *types.GetCurrentUserInfoReq) (*types.GetCurrentUserInfoResp, error) {
	current := middleware.GetCurrent(ctx)
	if current == nil || current.User == nil {
		return nil, errs.NewBusiness("invalid credentials").WithReason(errs.ReasonUnauthorized)
	}

	user, err := u.userUseCase.GetByID(ctx, current.User.ID)
	if err != nil {
		return nil, err
	}

	return &types.GetCurrentUserInfoResp{
		ID:          user.ID,
		RoleSlug:    user.Role,
		Permissions: []string{}, // TODO
	}, nil
}
