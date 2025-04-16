package permission

import "context"

// Builder provides methods to build the permission tree
type Builder struct{}

// NewBuilder creates a new permission tree builder
func NewBuilder() *Builder {
	return &Builder{}
}

// GetTree returns the permission tree
func (b *Builder) GetTree(ctx context.Context) ([]*Node, error) {
	// Define all permissions, grouped by module
	permissions := []*Node{
		{
			ID:   "menu:home",
			Name: "Home",
		},
		{
			ID:   "menu:setting",
			Name: "System Management",
			Children: []*Node{
				{
					ID:   "menu:department",
					Name: "Department Management",
					Children: []*Node{
						{ID: DepartmentList, Name: "Department List"},
						{ID: DepartmentDetail, Name: "Department Detail"},
						{ID: DepartmentCreate, Name: "Create Department"},
						{ID: DepartmentUpdate, Name: "Update Department"},
						{ID: DepartmentUpdateEnabled, Name: "Update Department Enabled"},
						{ID: DepartmentDelete, Name: "Delete Department"},
					},
				},
				{
					ID:   "menu:role",
					Name: "Role Management",
					Children: []*Node{
						{ID: RoleList, Name: "Role List"},
						{ID: RoleDetail, Name: "Role Detail"},
						{ID: RoleCreate, Name: "Create Role"},
						{ID: RoleUpdate, Name: "Update Role"},
						{ID: RoleUpdateEnabled, Name: "Update Role Enabled"},
						{ID: RoleDelete, Name: "Delete Role"},
						{ID: PermissionUpdate, Name: "Set Permissions"},
					},
				},
				{
					ID:   "menu:user",
					Name: "User Management",
					Children: []*Node{
						{ID: UserList, Name: "User List"},
						{ID: UserDetail, Name: "User Detail"},
						{ID: UserCreate, Name: "Create User"},
						{ID: UserUpdate, Name: "Update User"},
						{ID: UserUpdateEnabled, Name: "Update User Enabled"},
						{ID: UserDelete, Name: "Delete User"},
					},
				},
			},
		},
	}

	return permissions, nil
}
