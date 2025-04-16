package types

const (
	DefaultPage     = 1
	DefaultPageSize = 10
)

// PageReq represents common pagination parameters for list query requests.
type PageReq struct {
	Page     int `form:"page" json:"page" binding:"omitempty,min=1"`                   // Current page number (1-based)
	PageSize int `form:"page_size" json:"page_size" binding:"omitempty,min=1,max=100"` // Items per page
}

// PageResp represents pagination metadata in responses.
type PageResp struct {
	Page       int `json:"page"`        // Current page number
	PageSize   int `json:"page_size"`   // Items per page
	TotalCount int `json:"total_count"` // Total number of items
	TotalPages int `json:"total_pages"` // Total number of pages
}

// GetOffset calculates the offset for database queries.
func (p *PageReq) GetOffset() int {
	return (p.GetPage() - 1) * p.GetPageSize()
}

// GetLimit returns the limit for database queries.
func (p *PageReq) GetLimit() int {
	return p.GetPageSize()
}

// GetPage returns the current page, defaulting to 1 if not set.
func (p *PageReq) GetPage() int {
	if p.Page <= 0 {
		return DefaultPage
	}
	return p.Page
}

// GetPageSize returns the page size, defaulting to 10 if not set.
func (p *PageReq) GetPageSize() int {
	if p.PageSize <= 0 {
		return DefaultPageSize
	}
	return p.PageSize
}

// NewPageResp creates pagination metadata.
func NewPageResp(count, page, pageSize int) PageResp {
	if page <= 0 {
		page = DefaultPage
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}

	pages := count / pageSize
	if count%pageSize > 0 {
		pages++
	}

	return PageResp{
		TotalCount: count,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: pages,
	}
}
