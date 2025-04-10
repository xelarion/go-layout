package types

import (
	"regexp"
	"strings"
)

// Sort direction constants
const (
	SortAsc  = "asc"
	SortDesc = "desc"
)

// Pattern allows alphanumeric characters, underscores and dots (for table prefixes)
// This pattern prevents SQL injection by rejecting SQL metacharacters
var safeSortFieldPattern = regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)

// SortReq represents common sorting parameters for list query requests.
type SortReq struct {
	SortField string `form:"sort_field" json:"sort_field" binding:"omitempty"`                         // Field to sort by, can include table prefix (e.g. "users.id")
	SortOrder string `form:"sort_order" json:"sort_order" binding:"omitempty,oneof=asc desc ASC DESC"` // Sort direction: asc or desc
}

// GetSortField returns the field to sort by or empty string if not set.
func (s *SortReq) GetSortField() string {
	return s.SortField
}

// GetSortOrder returns the sort direction (asc/desc), defaulting to asc if not set.
func (s *SortReq) GetSortOrder() string {
	order := strings.ToLower(s.SortOrder)
	if order == SortDesc {
		return SortDesc
	}
	return SortAsc
}

// GetSortClause returns a properly formatted "field direction" string for SQL ORDER BY
// or empty string if sort field is not set or invalid.
func (s *SortReq) GetSortClause() string {
	if s.GetSortField() == "" {
		return ""
	}

	// Validate field name with regex pattern
	if !safeSortFieldPattern.MatchString(s.GetSortField()) {
		return ""
	}

	return s.GetSortField() + " " + s.GetSortOrder()
}
