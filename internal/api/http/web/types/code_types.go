package types

// GenerateCodeReq represents code generation request.
type GenerateCodeReq struct {
	RuleSlug string `json:"rule_slug" binding:"required"` // 编码规则标识
}

// GenerateCodeResp represents code generation response.
type GenerateCodeResp struct {
	Code string `json:"code"` // 生成的编码
}

// PreviewCodeReq represents code preview request.
type PreviewCodeReq struct {
	Prefix        string `json:"prefix" binding:"required"`         // 编码前缀
	TimeRule      string `json:"time_rule" binding:"required"`      // 时间规则
	GenerateType  string `json:"generate_type" binding:"required"`  // 生成类型
	InitialSerial int32  `json:"initial_serial" binding:"required"` // 初始流水号
	SerialLength  int32  `json:"serial_length" binding:"required"`  // 流水号长度
	Step          int32  `json:"step" binding:"required"`           // 步长
}

// PreviewCodeResp represents code preview response.
type PreviewCodeResp struct {
	Code         string `json:"code"`          // 预览的编码
	GenerateRule string `json:"generate_rule"` // 生成规则
}
