package types

// Common code_rule-related types

// ListCodeRulesReq represents code_rule list query parameters.
type ListCodeRulesReq struct {
	PageReq
	SortReq
	Key          string `form:"key" binding:"omitempty,max=100"`   // 搜索关键词
	GenerateType string `form:"generate_type" binding:"omitempty"` // 生成类型
	TargetForm   string `form:"target_form" binding:"omitempty"`   // 目标表单
}

// ListCodeRulesResp represents code_rule list with pagination info.
type ListCodeRulesResp struct {
	Results    []ListCodeRulesRespResult `json:"results"`
	Pagination PageResp                  `json:"pagination"`
}

// ListCodeRulesRespResult represents a single code_rule in the list.
type ListCodeRulesRespResult struct {
	ID            uint   `json:"id"`
	TargetForm    string `json:"target_form"`    // 目标表单
	Name          string `json:"name"`           // 编码名称
	Prefix        string `json:"prefix"`         // 编号前缀
	TimeRule      string `json:"time_rule"`      // 时间规则
	GenerateType  string `json:"generate_type"`  // 生成类型
	InitialSerial int32  `json:"initial_serial"` // 初始流水号
	SerialLength  int32  `json:"serial_length"`  // 流水号长度
	Step          int32  `json:"step"`           // 步长
	Preview       string `json:"preview"`        // 编号预览
	GenerateRule  string `json:"generate_rule"`  // 编码生成规则
}

type GetCodeRuleFormDataReq struct {
	ID uint `uri:"id" binding:"required" swaggerignore:"true"`
}

// GetCodeRuleFormDataResp represents data needed for code_rule forms (create/update).
type GetCodeRuleFormDataResp struct {
	ID            uint   `json:"id"`
	TargetForm    string `json:"target_form"`    // 目标表单
	Name          string `json:"name"`           // 编码名称
	Prefix        string `json:"prefix"`         // 编号前缀
	TimeRule      string `json:"time_rule"`      // 时间规则
	GenerateType  string `json:"generate_type"`  // 生成类型
	InitialSerial int32  `json:"initial_serial"` // 初始流水号
	SerialLength  int32  `json:"serial_length"`  // 流水号长度
	Step          int32  `json:"step"`           // 步长
	Preview       string `json:"preview"`        // 编号预览
	GenerateRule  string `json:"generate_rule"`  // 编码生成规则
}

// UpdateCodeRuleReq represents code_rule update data.
type UpdateCodeRuleReq struct {
	ID            uint   `uri:"id" binding:"required" swaggerignore:"true"`
	Prefix        string `json:"prefix" binding:"required,min=1,max=100"` // 编号前缀
	TimeRule      string `json:"time_rule" binding:"required"`            // 时间规则 slug
	GenerateType  string `json:"generate_type" binding:"required"`        // 生成类型 slug
	InitialSerial int32  `json:"initial_serial" binding:"required"`       // 初始流水号
	SerialLength  int32  `json:"serial_length" binding:"required"`        // 流水号长度
}

// UpdateCodeRuleResp represents code_rule update response.
type UpdateCodeRuleResp struct {
}

type GetCodeRuleTargetFormOptionsReq struct {
	SortReq
}

type GetCodeRuleTargetFormOptionsResp struct {
	Results []GetCodeRuleTargetFormOptionsRespResult `json:"results"`
}

type GetCodeRuleTargetFormOptionsRespResult struct {
	GetOptionsRespResult
}
