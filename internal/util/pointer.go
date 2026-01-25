package util

// PtrValue returns the value of the pointer, or the default value if the pointer is nil.
func PtrValue[T any](p *T, def T) T {
	if p == nil {
		return def
	}
	return *p
}

// PtrEqual returns true if the two pointers are equal, false otherwise.
func PtrEqual[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
