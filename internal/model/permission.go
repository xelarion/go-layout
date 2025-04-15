package model

// Permission represents a system permission
type Permission struct {
	ID       string        // Permission identifier, like "user:list"
	Name     string        // Permission name
	Children []*Permission // Child permissions
}
