package types

type (
	// CommonResponseDTO 通用的返回报文结构体
	CommonResponseDTO struct {
		Code         int         `json:"code"`
		Data         interface{} `json:"data,omitempty"`
		ErrorMessage string      `json:"err_msg,omitempty"`
	}

	// RuleDTO Rule的HTTP报文结构
	RuleDTO struct {
		ID          string           `json:"id,omitempty"`
		Path        string           `json:"path,omitempty"`
		Method      string           `json:"method,omitempty"`
		Variable    VariableDTO      `json:"variable,omitempty"`
		Weight      WeightDTO        `json:"weight,omitempty"`
		Regulations []*RegulationDTO `json:"responses,omitempty"`
	}

	// VariableDTO 变量的HTTP报文结构
	VariableDTO map[string]interface{}

	// WeightDTO 权重值的HTTP报文结构
	WeightDTO map[string]map[string]uint

	// RegulationDTO 响应报文规则的结构
	RegulationDTO struct {
		IsDefault bool         `json:"is_default,omitempty"`
		Filter    *FilterDTO   `json:"filter,omitempty"`
		Template  *TemplateDTO `json:"response,omitempty"`
	}

	// FilterDTO 筛选器的HTTP报文结构
	FilterDTO struct {
		Header map[string]string `json:"header,omitempty"`
		Query  map[string]string `json:"query,omitempty"`
		Body   map[string]string `json:"body,omitempty"`
	}

	// TemplateDTO 模板的HTTP报文结构
	TemplateDTO struct {
		IsTemplate       bool              `json:"is_template,omitempty"`
		IsHeaderTemplate bool              `json:"is_header_template,omitempty"`
		Header           map[string]string `json:"header,omitempty"`
		StatusCode       int               `json:"status_code,omitempty"`
		Body             string            `json:"body,omitempty"`
		B64EncodeBody    string            `json:"base64encoded_body,omitempty"`
	}
)
