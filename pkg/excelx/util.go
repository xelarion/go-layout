package excelx

// CellValue returns the value of the pointer, or "" if the pointer is nil.
func CellValue[T any](p *T) any {
	if p == nil {
		return ""
	}
	return *p
}
