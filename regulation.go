package deepmock

import (
	"encoding/base64"
	"html/template"
	"io"

	"github.com/qastub/deepmock/types"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type (
	// responseRegulation http响应报文渲染规则
	responseRegulation struct {
		isDefault        bool
		requestFilter    *requestFilter
		responseTemplate *responseTemplate
	}

	// responseTemplate http响应报文模板
	responseTemplate struct {
		isTemplate   bool
		isBinData    bool
		header       *fasthttp.ResponseHeader
		body         []byte
		htmlTemplate *template.Template
		raw          *types.ResourceResponseTemplate
	}
)

func newResponseTemplate(rrt *types.ResourceResponseTemplate) (*responseTemplate, error) {
	var body []byte
	var err error
	var isBin bool
	if rrt.B64EncodedBody != "" {
		isBin = true
		body, err = base64.StdEncoding.DecodeString(rrt.B64EncodedBody)
		if err != nil {
			Logger.Error("failed to decode base64encoded body data", zap.Error(err))
			return nil, err
		}
	} else {
		body = []byte(rrt.Body)
	}

	header := new(fasthttp.ResponseHeader)
	header.SetStatusCode(rrt.StatusCode)
	for k, v := range rrt.Header {
		header.Set(k, v)
	}

	rt := &responseTemplate{
		isTemplate: rrt.IsTemplate,
		isBinData:  isBin,
		header:     header,
		body:       body,
		raw:        rrt,
	}

	if rt.isTemplate {
		tmpl, err := template.New(genRandomString(16)).Parse(string(rt.body))
		if err != nil {
			Logger.Error("failed to parse html template", zap.ByteString("template", rt.body), zap.Error(err))
			return nil, err
		}
		rt.htmlTemplate = tmpl
	}
	return rt, nil
}

func (rt *responseTemplate) renderTemplate(rc renderContext, w io.Writer) error {
	if !rt.isTemplate {
		return nil
	}
	return rt.htmlTemplate.Execute(w, rc)
}

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
