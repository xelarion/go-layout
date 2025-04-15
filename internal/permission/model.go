package permission

// Node represents a system permission node
type Node struct {
	ID       string  `json:"id"`                 // Permission identifier, like "user:list"
	Name     string  `json:"name"`               // Permission name
	Children []*Node `json:"children,omitempty"` // Child permissions
}

// TreeNode represents a node in the permission tree
type TreeNode struct {
	Node
}
