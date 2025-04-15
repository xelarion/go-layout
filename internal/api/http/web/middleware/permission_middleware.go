package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xelarion/go-layout/internal/api/http/web/types"
	"github.com/xelarion/go-layout/internal/usecase"
)

// PermissionMiddleware handles permission checking
type PermissionMiddleware struct {
	roleUseCase *usecase.RoleUseCase
	logger      *zap.Logger
}

// NewPermissionMiddleware creates a new instance of PermissionMiddleware
func NewPermissionMiddleware(roleUseCase *usecase.RoleUseCase, logger *zap.Logger) *PermissionMiddleware {
	return &PermissionMiddleware{
		roleUseCase: roleUseCase,
		logger:      logger.Named("permission_middleware"),
	}
}

// Check checks if the user has any of the required permissions.
// If only one permission is provided, it checks for that specific permission.
// If multiple permissions are provided, it checks if the user has any of them.
func (m *PermissionMiddleware) Check(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get current context
		current := GetCurrent(c.Request.Context())
		if current == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.Error(types.CodeUnauthorized, "Unauthorized"))
			return
		}

		// Get role
		role, err := m.roleUseCase.GetByID(c.Request.Context(), current.RoleID)
		if err != nil {
			m.logger.Error("Failed to get role", zap.Error(err), zap.Uint("role_id", current.RoleID))
			c.AbortWithStatusJSON(http.StatusInternalServerError, types.Error(types.CodeInternalError, "Server error"))
			return
		}

		// If only one permission is provided, check for that specific permission
		if len(permissions) == 1 {
			if !role.HasPermission(permissions[0]) {
				c.AbortWithStatusJSON(http.StatusForbidden, types.Error(types.CodeForbidden, "Access denied"))
				return
			}
		} else if len(permissions) > 1 {
			// If multiple permissions are provided, check if the user has any of them
			if !role.HasAnyPermission(permissions...) {
				c.AbortWithStatusJSON(http.StatusForbidden, types.Error(types.CodeForbidden, "Access denied"))
				return
			}
		}

		c.Next()
	}
}
