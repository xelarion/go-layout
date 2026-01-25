package types

// Common user-related types

// CreateUserReq represents user creation request.
type CreateUserReq struct {
	FullName    string `json:"full_name" binding:"required,min=1,max=100"`   // 用户姓名
	EmployeeNo  string `json:"employee_no" binding:"required,min=1,max=50"`  // 工号
	Username    string `json:"username" binding:"required,min=1,max=100"`    // 用户名
	Gender      string `json:"gender" binding:"required"`                    // 性别
	PositionID  uint   `json:"position_id" binding:"required,min=1"`         // 岗位
	PhoneNumber string `json:"phone_number" binding:"required,min=1,max=20"` // 手机号
	Email       string `json:"email" binding:"required,email,max=100"`       // 邮箱
	RoleID      uint   `json:"role_id" binding:"required,min=1"`             // 角色
	Enabled     bool   `json:"enabled" binding:"omitempty"`                  // 状态
	Remark      string `json:"remark" binding:"omitempty,max=1000"`          // 备注
}

// CreateUserResp represents user creation response.
type CreateUserResp struct {
	ID uint `json:"id"`
}

// ListUsersReq represents user list query parameters.
type ListUsersReq struct {
	PageReq
	SortReq
	Key            string `form:"key" binding:"omitempty,max=100"` // Search by username or full name
	Enabled        *bool  `form:"enabled"`
	RoleID         uint   `form:"role_id" binding:"omitempty,gte=1"`
	OrganizationID uint   `form:"organization_id" binding:"omitempty,gte=1"`
}

// ListUsersResp represents user list with pagination info.
type ListUsersResp struct {
	Results    []ListUsersRespResult `json:"results"`
	Pagination PageResp              `json:"pagination"`
}

type ListUsersRespResult struct {
	ID               uint   `json:"id"`
	FullName         string `json:"full_name"`         // 用户姓名
	EmployeeNo       string `json:"employee_no"`       // 工号
	Username         string `json:"username"`          // 用户名
	Gender           string `json:"gender"`            // 性别
	OrganizationName string `json:"organization_name"` // 组织名称
	PositionName     string `json:"position_name"`     // 岗位名称
	PhoneNumber      string `json:"phone_number"`      // 手机号
	Email            string `json:"email"`             // 邮箱
	RoleName         string `json:"role_name"`         // 角色名称
	RoleSlug         string `json:"role_slug"`         // 角色标识
	Enabled          bool   `json:"enabled"`           // 状态
}

// GetUserReq represents a request to get a specific user.
type GetUserReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetUserResp represents a user object in responses.
type GetUserResp struct {
	ID               uint   `json:"id"`                // ID
	CreatedAt        string `json:"created_at"`        // 创建时间
	FullName         string `json:"full_name"`         // 用户姓名
	EmployeeNo       string `json:"employee_no"`       // 工号
	Username         string `json:"username"`          // 用户名
	Gender           string `json:"gender"`            // 性别
	OrganizationName string `json:"organization_name"` // 组织名称
	PositionName     string `json:"position_name"`     // 岗位名称
	PhoneNumber      string `json:"phone_number"`      // 手机号
	Email            string `json:"email"`             // 邮箱
	RoleName         string `json:"role_name"`         // 角色名称
	RoleSlug         string `json:"role_slug"`         // 角色标识
	Enabled          bool   `json:"enabled"`           // 状态
}

type GetUserFormDataReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetUserFormDataResp represents data needed for user forms (create/update).
type GetUserFormDataResp struct {
	ID          uint   `json:"id"`
	FullName    string `json:"full_name"`    // 用户姓名
	EmployeeNo  string `json:"employee_no"`  // 工号
	Username    string `json:"username"`     // 用户名
	Gender      string `json:"gender"`       // 性别
	PositionID  uint   `json:"position_id"`  // 岗位
	PhoneNumber string `json:"phone_number"` // 手机号
	Email       string `json:"email"`        // 邮箱
	RoleID      uint   `json:"role_id"`      // 角色
	RoleSlug    string `json:"role_slug"`    // 角色标识
	Enabled     bool   `json:"enabled"`      // 状态
	Remark      string `json:"remark"`       // 备注
}

// UpdateUserReq represents user update data.
type UpdateUserReq struct {
	ID          uint   `uri:"id" binding:"required" swaggerignore:"true"`
	FullName    string `json:"full_name" binding:"required,min=1,max=100"`   // 用户姓名
	EmployeeNo  string `json:"employee_no" binding:"required,min=1,max=50"`  // 工号
	Username    string `json:"username" binding:"required,min=1,max=100"`    // 用户名
	Gender      string `json:"gender" binding:"required"`                    // 性别
	PositionID  uint   `json:"position_id" binding:"required,min=1"`         // 岗位
	PhoneNumber string `json:"phone_number" binding:"required,min=1,max=20"` // 手机号
	Email       string `json:"email" binding:"required,email,max=100"`       // 邮箱
	RoleID      uint   `json:"role_id" binding:"required,min=1"`             // 角色
	Enabled     bool   `json:"enabled" binding:"omitempty"`                  // 状态
	Remark      string `json:"remark" binding:"omitempty,max=1000"`          // 备注
}

// UpdateUserResp represents user update response.
type UpdateUserResp struct {
}

// UpdateUserEnabledReq represents a request to update user enabled status.
type UpdateUserEnabledReq struct {
	ID      uint  `uri:"id" binding:"required" swaggerignore:"true"`
	Enabled *bool `json:"enabled" binding:"required"`
}

// UpdateUserEnabledResp represents user update enabled status response.
type UpdateUserEnabledResp struct {
}

// UpdateUserPasswordReq represents a request to update user password.
type UpdateUserPasswordReq struct {
	ID       uint   `uri:"id" binding:"required" swaggerignore:"true"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// UpdateUserPasswordResp represents user update password response.
type UpdateUserPasswordResp struct {
}

// DeleteUserReq represents a request to delete a user.
type DeleteUserReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// DeleteUserResp represents user delete response.
type DeleteUserResp struct {
}

// BatchDeleteUsersReq represents a request to batch delete users.
type BatchDeleteUsersReq struct {
	IDs []uint `json:"ids" binding:"required,min=1,dive,required"`
}

// BatchDeleteUsersResp represents batch delete users response.
type BatchDeleteUsersResp struct {
	SuccessCount int      `json:"success_count"` // 成功删除数量
	FailedCount  int      `json:"failed_count"`  // 失败数量
	FailedItems  []string `json:"failed_items"`  // 失败的信息（格式：名称: 错误原因）
}

type GetUserOptionsReq struct {
	SortReq
	Enabled *bool `form:"enabled"` // 状态筛选
}

// GetUserOptionsResp represents user options response.
type GetUserOptionsResp struct {
	Results []GetUserOptionsRespResult `json:"results"`
}

type GetUserOptionsRespResult struct {
	GetOptionsRespResult
	Enabled bool `json:"enabled"`
}

// ExportUsersReq represents user export request.
type ExportUsersReq struct {
	IDs            []uint `json:"ids" binding:"omitempty"`                   // 指定导出的用户ID列表
	Key            string `json:"key" binding:"omitempty,max=100"`           // 搜索关键字
	Enabled        *bool  `json:"enabled"`                                   // 状态筛选
	RoleID         uint   `json:"role_id" binding:"omitempty,gte=1"`         // 角色筛选
	OrganizationID uint   `json:"organization_id" binding:"omitempty,gte=1"` // 组织筛选
}
