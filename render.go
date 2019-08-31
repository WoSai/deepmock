package deepmock

import (
	"bytes"
	"net/url"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

var (
	formContentType = []byte("application/x-www-form-urlencoded")
	jsonContentType = []byte("application/json")
)

type (
	renderContext struct {
		Variable ruleVariable
		Weight   params
		Header   params
		Query    params
		Form     params
		Json     map[string]interface{}
	}

	params map[string]string
)

func extractHeaderAsParams(req *fasthttp.Request) params {
	p := make(params)
	req.Header.VisitAll(func(key, value []byte) {
		p[string(key)] = string(value)
	})
	return p
}

func extractQueryAsParams(req *fasthttp.Request) params {
	p := make(params)
	req.URI().QueryArgs().VisitAll(func(key, value []byte) {
		p[string(key)] = string(value)
	})
	return p
}

func extractBodyAsParams(req *fasthttp.Request) (params, map[string]interface{}) {
	ct := req.Header.ContentType()

	switch {
	case bytes.Contains(ct, formContentType):
		val, err := url.ParseQuery(string(req.Body()))
		if err != nil {
			Logger.Error("failed to automatic parse form data", zap.Error(err))
		}
		f := make(params)
		for k := range val {
			f[k] = val.Get(k)
		}
		return f, nil

	case bytes.Contains(ct, jsonContentType):
		j := make(map[string]interface{})
		err := json.Unmarshal(req.Body(), &j)
		if err != nil {
			Logger.Error("failed to automatic unmarshal request body", zap.Error(err))
			return nil, nil
		}
		return nil, j

	default:
		return nil, nil
	}
}
