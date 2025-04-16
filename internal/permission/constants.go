package permission

// Permission constants
const (
	// All permissions
	All = "*"

	DepartmentList          = "department/list"    // Department list
	DepartmentDetail        = "department/detail"  // Department detail
	DepartmentCreate        = "department/create"  // Create department
	DepartmentUpdate        = "department/update"  // Update department
	DepartmentUpdateEnabled = "department/enabled" // Update department enabled
	DepartmentDelete        = "department/delete"  // Delete department

	RoleList          = "role/list"    // Role list
	RoleDetail        = "role/detail"  // Role detail
	RoleCreate        = "role/create"  // Create role
	RoleUpdate        = "role/update"  // Update role
	RoleUpdateEnabled = "role/enabled" // Update role enabled
	RoleDelete        = "role/delete"  // Delete role

	PermissionUpdate = "permission/update" // Update permission\

	UserList          = "user/list"    // User list
	UserDetail        = "user/detail"  // User detail
	UserCreate        = "user/create"  // Create user
	UserUpdate        = "user/update"  // Update user
	UserUpdateEnabled = "user/enabled" // Update user enabled
	UserDelete        = "user/delete"  // Delete user
)
