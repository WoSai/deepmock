package types

type (
	CommonResponseDTO struct {
		Code         int         `json:"code"`
		Data         interface{} `json:"data,omitempty"`
		ErrorMessage string      `json:"err_msg,omitempty"`
	}

	RuleDTO struct {
		ID       string           `json:"id,omitempty"`
		Path     string           `json:"path,omitempty"`
		Method   string           `json:"method,omitempty"`
		Variable VariableDTO      `json:"variable,omitempty"`
		Weight   WeightDTO        `json:"weight,omitempty"`
		Response []*RegulationDTO `json:"responses,omitempty"`
	}

	VariableDTO map[string]interface{}

	WeightDTO map[string]map[string]uint

	RegulationDTO struct {
		IsDefault bool         `json:"is_default,omitempty"`
		Filter    *FilterDTO   `json:"filter,omitempty"`
		Template  *TemplateDTO `json:"response,omitempty"`
	}

	FilterDTO struct {
		Header map[string]string `json:"header,omitempty"`
		Query  map[string]string `json:"query,omitempty"`
		Body   map[string]string `json:"body,omitempty"`
	}

	TemplateDTO struct {
		IsTemplate    bool              `json:"is_template,omitempty"`
		Header        map[string]string `json:"header,omitempty"`
		StatusCode    int               `json:"status_code,omitempty"`
		Body          string            `json:"body,omitempty"`
		B64EncodeBody string            `json:"base64encoded_body,omitempty"`
	}
)
