package domain

import (
	"bytes"

	"github.com/goccy/go-json"
	"github.com/valyala/fasthttp"
)

var (
	formContentType      = []byte("application/x-www-form-urlencoded")
	multipartContentType = []byte("multipart/form-data")
	jsonContentType      = []byte("application/json")
)

func extractHeaderAsParams(req *fasthttp.Request) map[string]string {
	p := make(map[string]string)
	req.Header.VisitAll(func(key, value []byte) {
		p[string(key)] = string(value)
	})
	return p
}

func extractQueryAsParams(req *fasthttp.Request) map[string]string {
	p := make(map[string]string)
	req.URI().QueryArgs().VisitAll(func(key, value []byte) {
		p[string(key)] = string(value)
	})
	return p
}

func extractBodyAsParams(req *fasthttp.Request) (map[string]string, map[string]interface{}) {
	ct := req.Header.ContentType()

	switch {
	case bytes.HasPrefix(ct, formContentType):
		p := make(map[string]string)
		req.PostArgs().VisitAll(func(key, value []byte) {
			p[string(key)] = string(value)
		})
		return p, nil

	case bytes.HasPrefix(ct, multipartContentType):
		p := make(map[string]string)
		form, err := req.MultipartForm()
		if err != nil {
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
			return nil, nil
		}
		return nil, j

	default:
		return nil, nil
	}
}
