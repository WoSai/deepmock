package deepmock

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"regexp"
	"sync"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type (
	ruleManager struct {
		executors     map[string]*ruleExecutor
		executorCache simplelru.LRUCache
		mu            sync.RWMutex
	}

	ruleExecutor struct {
		requestMatcher  *requestMatcher
		context         ResourceContext
		weightingFactor ResourceWeightingFactor
		mockResponses   []*responseRegulation
	}

	// requestMatcher 请求匹配器
	requestMatcher struct {
		id     string
		path   []byte
		method []byte
		re     *regexp.Regexp
		raw    *ResourceRequestMatcher
	}

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
		raw          *ResourceResponseTemplate
	}
)

func newRequestMatcher(res *ResourceRequestMatcher) (*requestMatcher, error) {
	if err := res.check(); err != nil {
		return nil, err
	}
	re, err := regexp.Compile(res.Path)
	if err != nil {
		Logger.Error("failed to compile regular expression", zap.String("path", res.Path), zap.Error(err))
		return nil, err
	}

	rm := &requestMatcher{
		path:   []byte(res.Path),
		method: bytes.ToUpper([]byte(res.Method)),
		re:     re,
		raw:    res,
	}
	rm.id = genID(rm.path, rm.method)
	return rm, nil
}

func (rm *requestMatcher) match(path, method []byte) bool {
	if bytes.Compare(rm.method, method) != 0 {
		return false
	}
	return rm.re.Match(path)
}

func newResponseTemplate(rrt *ResourceResponseTemplate) (*responseTemplate, error) {
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
		tmpl, err := template.New("template-name").Parse(string(rt.body))
		if err != nil {
			Logger.Error("failed to parse html template", zap.ByteString("template", rt.body), zap.Error(err))
			return nil, err
		}
		rt.htmlTemplate = tmpl
	}
	return rt, nil
}

func (mr *responseRegulation) filter(req *fasthttp.Request) bool {
	return mr.requestFilter.filter(req)
}

func (re *ruleExecutor) id() string {
	return re.requestMatcher.id
}
