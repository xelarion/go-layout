package middleware

import (
	appctx "github.com/xelarion/go-layout/internal/context"
)

var (
	NewCurrent = appctx.NewCurrent
	SetCurrent = appctx.SetCurrent
	GetCurrent = appctx.GetCurrent
)
