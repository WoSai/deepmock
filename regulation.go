package deepmock

import (
	"github.com/valyala/fasthttp"
	"github.com/wosai/deepmock/types"
)

type (
	// responseRegulation http响应报文渲染规则
	responseRegulation struct {
		isDefault        bool
		requestFilter    *requestFilter
		responseTemplate *responseTemplate
	}
)

func newResponseRegulation(res *types.ResourceResponseRegulation) (*responseRegulation, error) {
	if err := res.Check(); err != nil {
		return nil, err
	}

	rr := new(responseRegulation)
	rr.isDefault = res.IsDefault

	rf, err := newRequestFilter(res.Filter)
	if err != nil {
		return nil, err
	}
	rr.requestFilter = rf

	rt, err := newResponseTemplate(res.Response)
	if err != nil {
		return nil, err
	}
	rr.responseTemplate = rt
	return rr, nil
}

func (mr *responseRegulation) filter(req *fasthttp.Request) bool {
	return mr.requestFilter.filter(req)
}

func (mr *responseRegulation) wrap() *types.ResourceResponseRegulation {
	rrr := new(types.ResourceResponseRegulation)
	rrr.IsDefault = mr.isDefault
	rrr.Response = mr.responseTemplate.raw
	rrr.Filter = mr.requestFilter.wrap()
	return rrr
}
