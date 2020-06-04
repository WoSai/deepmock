package mock

import (
	"bytes"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"github.com/wosai/deepmock"
	"go.uber.org/zap"
)

var (
	json                 = jsoniter.ConfigCompatibleWithStandardLibrary
	formContentType      = []byte("application/x-www-form-urlencoded")
	multipartContentType = []byte("multipart/form-data")
	jsonContentType      = []byte("application/json")
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
	case bytes.HasPrefix(ct, formContentType):
		p := make(params)
		req.PostArgs().VisitAll(func(key, value []byte) {
			p[string(key)] = string(value)
		})
		return p, nil

	case bytes.HasPrefix(ct, multipartContentType):
		p := make(params)
		form, err := req.MultipartForm()
		if err != nil {
			deepmock.Logger.Error("bad multipart form data", zap.Error(err))
			return nil, nil
		}
		for k, v := range form.Value {
			p[k] = v[0]
		}
		return p, nil

	case bytes.HasPrefix(ct, jsonContentType):
		j := make(map[string]interface{})
		err := json.Unmarshal(req.Body(), &j)
		if err != nil {
			deepmock.Logger.Error("failed to automatic unmarshal request body", zap.Error(err))
			return nil, nil
		}
		return nil, j

	default:
		return nil, nil
	}
}
