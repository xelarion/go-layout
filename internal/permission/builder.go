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
			ID:   "menu:system",
			Name: "System Management",
			Children: []*Node{
				{
					ID:   "menu:user",
					Name: "User Management",
					Children: []*Node{
						{ID: UserList, Name: "User List"},
						{ID: UserDetail, Name: "User Detail"},
						{ID: UserCreate, Name: "Create User"},
						{ID: UserUpdate, Name: "Update User"},
						{ID: UserDelete, Name: "Delete User"},
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
						{ID: RoleDelete, Name: "Delete Role"},
					},
				},
				{
					ID:   "menu:department",
					Name: "Department Management",
					Children: []*Node{
						{ID: DepartmentList, Name: "Department List"},
						{ID: DepartmentDetail, Name: "Department Detail"},
						{ID: DepartmentCreate, Name: "Create Department"},
						{ID: DepartmentUpdate, Name: "Update Department"},
						{ID: DepartmentDelete, Name: "Delete Department"},
					},
				},
				{
					ID:   "menu:permission",
					Name: "Permission Management",
					Children: []*Node{
						{ID: PermissionList, Name: "Permission List"},
						{ID: PermissionUpdate, Name: "Set Permissions"},
					},
				},
			},
		},
	}

	return permissions, nil
}
