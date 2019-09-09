package types

import (
	"errors"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type (
	CommonResource struct {
		Code         int         `json:"code"`
		Data         interface{} `json:"data,omitempty"`
		ErrorMessage string      `json:"err_msg,omitempty"`
	}

	ResourceRequestMatcher struct {
		Path   string `json:"path"`
		Method string `json:"method"`
	}

	ResourceRule struct {
		ID           string                         `json:"id,omitempty" ddb:"id"`
		Path         string                         `json:"path,omitempty" ddb:"path"`
		Method       string                         `json:"method,omitempty" ddb:"method"`
		Variable     *ResourceVariable              `json:"variable,omitempty" ddb:"variable"`
		Weight       *ResourceWeight                `json:"weight,omitempty" ddb:"weight"`
		Responses    *ResourceResponseRegulationSet `json:"responses,omitempty" ddb:"responses"`
		Version      int                            `json:"-" ddb:"version"`
		CreatedTime  time.Time                      `json:"-" ddb:"ctime"`
		ModifiedTime time.Time                      `json:"-" ddb:"mtime"`
		Disabled     bool                           `json:"-" ddb:"disabled"`
	}

	ResourceResponseRegulation struct {
		IsDefault bool                      `json:"is_default,omitempty"`
		Filter    *ResourceFilter           `json:"filter,omitempty"`
		Response  *ResourceResponseTemplate `json:"response"`
	}

	ResourceVariable map[string]interface{}

	ResourceWeight map[string]ResourceWeightingFactor

	ResourceFilter struct {
		Header ResourceHeaderFilterParameters `json:"header,omitempty"`
		Query  ResourceQueryFilterParameters  `json:"query,omitempty"`
		Body   ResourceBodyFilterParameters   `json:"body,omitempty"`
	}

	ResourceResponseTemplate struct {
		IsTemplate     bool                   `json:"is_template,omitempty"`
		Header         ResourceHeaderTemplate `json:"header,omitempty"`
		StatusCode     int                    `json:"status_code,omitempty"`
		Body           string                 `json:"body,omitempty"`
		B64EncodedBody string                 `json:"base64encoded_body,omitempty"`
	}

	ResourceHeaderFilterParameters map[string]string

	ResourceBodyFilterParameters map[string]string

	ResourceQueryFilterParameters map[string]string

	ResourceHeaderTemplate map[string]string

	ResourceWeightingFactor map[string]uint

	ResourceResponseRegulationSet []*ResourceResponseRegulation
)

func (rv *ResourceVariable) UnmarshalByte(data []byte) error {
	return json.Unmarshal(data, rv)
}

func (rw *ResourceWeight) UnmarshalByte(data []byte) error {
	return json.Unmarshal(data, rw)
}

func (rrr *ResourceResponseRegulationSet) UnmarshalByte(data []byte) error {
	return json.Unmarshal(data, rrr)
}

func (rmr *ResourceResponseRegulation) Check() error {
	if !rmr.IsDefault && rmr.Filter == nil {
		return errors.New("missing filter rule, or set as default response")
	}
	return nil
}

func (mrs ResourceResponseRegulationSet) Check() error {
	var d int
	if mrs == nil {
		return errors.New("missing mock response")
	}

	for _, r := range mrs {
		if r.IsDefault {
			d++
		}
		if err := r.Check(); err != nil {
			return err
		}
	}
	if d != 1 {
		return errors.New("no default response or provided more than one")
	}
	return nil
}

func (hfp ResourceHeaderFilterParameters) Check() error {
	if hfp == nil {
		return nil
	}

	if _, ok := hfp["mode"]; !ok {
		return errors.New("missing filter mode")
	}
	return nil
}

func (qfp ResourceQueryFilterParameters) Check() error {
	if qfp == nil {
		return nil
	}

	if _, ok := qfp["mode"]; !ok {
		return errors.New("missing filter mode")
	}
	return nil
}

func (bfp ResourceBodyFilterParameters) Check() error {
	if bfp == nil {
		return nil
	}

	if _, ok := bfp["mode"]; !ok {
		return errors.New("missing filter mode")
	}
	return nil
}
