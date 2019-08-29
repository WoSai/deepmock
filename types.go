package deepmock

import (
	"regexp"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type (
	Rule struct {
		Request   *RequestMatch   `json:"request"`
		Responses []*ResponseRule `json:"responses"`
		Context   RuleContext     `json:"context,omitempty"`
		Weights   RuleWeights     `json:"weights,omitempty"`
	}

	ResponseRule struct {
		Default  bool            `json:"default,omitempty"`
		Filter   *RequestFilter  `json:"filter,omitempty"`
		Response *ResponseRender `json:"response"`
	}

	RequestFilter struct {
		Headers HeaderFilterParameters `json:"headers,omitempty"`
		Body    BodyFilterParameters   `json:"body,omitempty"`
		Query   QueryFilterParameters  `json:"query,omitempty"`
	}

	RequestMatch struct {
		Path   string         `json:"path"`
		Method string         `json:"method"`
		re     *regexp.Regexp `json:"-"`
		once   sync.Once      `json:"-"`
	}

	ResponseRender struct {
		IsTemplate     bool           `json:"is_template,omitempty"`
		Headers        ResponseHeader `json:"headers,omitempty"`
		StatusCode     int            `json:"status_code,omitempty"`
		Body           string         `json:"body,omitempty"`
		B64EncodedBody string         `json:"base64encoded_body,omitempty"`
	}

	RuleWeights map[string]int

	RuleContext map[string]interface{}

	HeaderFilterParameters map[string]string

	BodyFilterParameters map[string]string

	QueryFilterParameters map[string]string

	ResponseHeader map[string]string
)
