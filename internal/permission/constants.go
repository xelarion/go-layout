package permission

// Permission constants
const (
	// All permissions
	All = "*"

	// User management
	UserList   = "user:list"   // User list
	UserDetail = "user:detail" // User detail
	UserCreate = "user:create" // Create user
	UserUpdate = "user:update" // Update user
	UserDelete = "user:delete" // Delete user

	// Role management
	RoleList   = "role:list"   // Role list
	RoleDetail = "role:detail" // Role detail
	RoleCreate = "role:create" // Create role
	RoleUpdate = "role:update" // Update role
	RoleDelete = "role:delete" // Delete role

	// Department management
	DepartmentList   = "department:list"   // Department list
	DepartmentDetail = "department:detail" // Department detail
	DepartmentCreate = "department:create" // Create department
	DepartmentUpdate = "department:update" // Update department
	DepartmentDelete = "department:delete" // Delete department

	// Permission management
	PermissionList   = "permission:list"   // Permission list
	PermissionUpdate = "permission:update" // Update permission
)
