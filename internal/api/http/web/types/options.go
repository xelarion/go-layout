package types

// GetOptionsResp represents common options response.
type GetOptionsResp struct {
	Results []GetOptionsRespResult `json:"results"`
}

// GetOptionsRespResult represents a single option.
type GetOptionsRespResult struct {
	Label string `json:"label"`
	Value any    `json:"value"`
}
