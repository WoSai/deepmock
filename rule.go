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
		requestMatcher      *requestMatcher
		context             ResourceContext
		weightPickers       map[string]*weightingPicker
		responseRegulations []*responseRegulation
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

	weightingPicker struct {
		total   uint
		storage []string
		raw     ResourceWeightingFactor
	}

	weightingFactorHub map[string]*weightingPicker
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

func (rm *requestMatcher) wrap() *ResourceRequestMatcher {
	return rm.raw
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
		tmpl, err := template.New(genRandomString(16)).Parse(string(rt.body))
		if err != nil {
			Logger.Error("failed to parse html template", zap.ByteString("template", rt.body), zap.Error(err))
			return nil, err
		}
		rt.htmlTemplate = tmpl
	}
	return rt, nil
}

func newResponseRegulation(res *ResourceResponseRegulation) (*responseRegulation, error) {
	if err := res.check(); err != nil {
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

func (mr *responseRegulation) wrap() *ResourceResponseRegulation {
	rrr := new(ResourceResponseRegulation)
	rrr.IsDefault = mr.isDefault
	rrr.Response = mr.responseTemplate.raw
	rrr.Filter = mr.requestFilter.wrap()
	return rrr
}

func newRuleExecutor(res *ResourceRule) (*ruleExecutor, error) {
	re := new(ruleExecutor)
	// init requestMatcher
	rm, err := newRequestMatcher(res.Request)
	if err != nil {
		return nil, err
	}
	rm.raw = res.Request

	// init context
	re.context = res.Context

	// init weight
	re.weightPickers = map[string]*weightingPicker{}
	for k, v := range res.Weight {
		re.weightPickers[k] = newWeighingPicker(v)
	}

	// init responseRegulation
	if err := res.Responses.check(); err != nil {
		return nil, err
	}
	re.responseRegulations = make([]*responseRegulation, len(res.Responses))
	for i, reg := range res.Responses {
		rr, err := newResponseRegulation(reg)
		if err != nil {
			return nil, err
		}
		re.responseRegulations[i] = rr
	}
	return re, nil
}

func (re *ruleExecutor) id() string {
	return re.requestMatcher.id
}

func newWeighingPicker(res ResourceWeightingFactor) *weightingPicker {
	wp := &weightingPicker{raw: res}
	wp.preDistribute()
	return wp
}

func (wp *weightingPicker) preDistribute() {
	wp.storage = []string{}
	for k, v := range wp.raw {
		for i := 0; i < int(v); i++ {
			wp.storage = append(wp.storage, k)
			wp.total++
		}
	}
}

func newWeightingFactorHub(res ResourceWeight) weightingFactorHub {
	wfh := make(weightingFactorHub)
	for k, v := range res {
		wfh[k] = newWeighingPicker(v)
	}
	return wfh
}

func (wfg weightingFactorHub) wrap() ResourceWeight {
	if wfg == nil {
		return nil
	}

	ret := make(ResourceWeight)
	for k, v := range wfg {
		ret[k] = v.raw
	}
	return ret
}
