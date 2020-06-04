package resource

import (
	"time"

	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

const (
	// FilterModeAlwaysTrue 总是通过
	FilterModeAlwaysTrue FilterMode = "always_true"
	// FilterModeExact key/value精确模式
	FilterModeExact FilterMode = "exact"
	// FilterModeKeyword 关键字模板，即确保contains(a, b)结果为true
	FilterModeKeyword FilterMode = "keyword"
	// FilterModeRegular 正则表达式模式
	FilterModeRegular FilterMode = "regular"
)

type (
	FilterMode = string

	CommonResource struct {
		Code         int         `json:"code"`
		Data         interface{} `json:"data,omitempty"`
		ErrorMessage string      `json:"err_msg,omitempty"`
	}

	//ResourceRequestMatcher struct {
	//	Path   string `json:"path"`
	//	Method string `json:"method"`
	//}

	// Rule 规则对象
	Rule struct {
		ID           string                `json:"id,omitempty"`
		Path         string                `json:"path,omitempty"`
		Method       string                `json:"method,omitempty"`
		Variable     Variable              `json:"variable,omitempty" `
		Weight       Weight                `json:"weight,omitempty"`
		Responses    ResponseRegulationSet `json:"responses,omitempty"`
		Version      int                   `json:"version,omitempty"`
		CreatedTime  time.Time             `json:"ctime,omitempty"`
		ModifiedTime time.Time             `json:"mtime,omitempty"`
		Disabled     bool                  `json:"disabled,omitempty"`
	}

	// ResponseRegulation mock response的规则
	ResponseRegulation struct {
		IsDefault bool              `json:"is_default,omitempty"`
		Filter    *Filter           `json:"filter,omitempty"`
		Response  *ResponseTemplate `json:"response"`
	}

	// Variable mock规则的自定义变量池
	Variable map[string]interface{}

	RequestMatcher struct {
		Path   string
		Method string
	}

	// Weight mock规则的权重变量池
	Weight map[string]WeightingFactor

	// Filter mock response的过滤器
	Filter struct {
		Header HeaderFilterParameters `json:"header,omitempty"`
		Query  QueryFilterParameters  `json:"query,omitempty"`
		Body   BodyFilterParameters   `json:"body,omitempty"`
	}

	// ResponseTemplate mock response的渲染模板
	ResponseTemplate struct {
		IsTemplate     bool           `json:"is_template,omitempty"`
		Header         HeaderTemplate `json:"header,omitempty"`
		StatusCode     int            `json:"status_code,omitempty"`
		Body           string         `json:"body,omitempty"`
		B64EncodedBody string         `json:"base64encoded_body,omitempty"`
	}

	// HeaderFilterParameters 请求头筛选器的参数
	HeaderFilterParameters map[string]string

	// BodyFilterParameters 请求body部分的筛选器
	BodyFilterParameters map[string]string

	// QueryFilterParameters query string的过滤参数
	QueryFilterParameters map[string]string

	// HeaderTemplate 请求头渲染模板
	HeaderTemplate map[string]string

	// WeightingFactor 权重因子
	WeightingFactor map[string]uint

	// ResponseRegulationSet mock response的规则集合
	ResponseRegulationSet []*ResponseRegulation
)
