package deepmock

import (
	"bytes"
	"encoding/base64"
	"errors"
	"html/template"
	"math/rand"
	"regexp"
	"sync"

	lru "github.com/hashicorp/golang-lru"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type (
	ruleManager struct {
		executors map[string]*ruleExecutor
		cache     *lru.ARCCache
		mu        sync.RWMutex
	}

	ruleExecutor struct {
		requestMatcher      *requestMatcher
		context             ResourceContext
		weightPicker        weightingPicker
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

	weightingDice struct {
		total   uint
		storage []string
		raw     ResourceWeightingFactor
	}

	weightingPicker map[string]*weightingDice
)

var (
	defaultRuleManager *ruleManager
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

func (rt *responseTemplate) render(rc renderContext, res *fasthttp.Response) error {
	if !rt.isTemplate {
		return nil
	}
	return rt.htmlTemplate.Execute(res.BodyWriter(), rc)
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
	re.requestMatcher = rm
	rm.raw = res.Request

	// init context
	re.context = res.Context

	// init weight
	re.weightPicker = map[string]*weightingDice{}
	for k, v := range res.Weight {
		re.weightPicker[k] = newWeighingPicker(v)
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

func (re *ruleExecutor) match(path, method []byte) bool {
	return re.requestMatcher.match(path, method)
}

func newWeighingPicker(res ResourceWeightingFactor) *weightingDice {
	wp := &weightingDice{raw: res}
	wp.preDistribute()
	return wp
}

func (wp *weightingDice) preDistribute() {
	wp.storage = []string{}
	for k, v := range wp.raw {
		for i := 0; i < int(v); i++ {
			wp.storage = append(wp.storage, k)
			wp.total++
		}
	}
}

func (wp *weightingDice) dice() string {
	return wp.storage[rand.Intn(int(wp.total))]
}

func newWeightingPicker(res ResourceWeight) weightingPicker {
	wfh := make(weightingPicker)
	for k, v := range res {
		wfh[k] = newWeighingPicker(v)
	}
	return wfh
}

func (wfg weightingPicker) wrap() ResourceWeight {
	if wfg == nil {
		return nil
	}

	ret := make(ResourceWeight)
	for k, v := range wfg {
		ret[k] = v.raw
	}
	return ret
}

func (wfg weightingPicker) dice() params {
	p := make(params)
	for k, w := range wfg {
		p[k] = w.dice()
	}
	return p
}

func newRuleManager() *ruleManager {
	cache, err := lru.NewARC(1000)
	if err != nil {
		panic(err)
	}

	return &ruleManager{
		executors: make(map[string]*ruleExecutor),
		cache:     cache,
	}
}

func (rm *ruleManager) genCacheID(path, method []byte) string {
	return string(method) + string(path)
}

func (rm *ruleManager) findExecutor(path, method []byte) (*ruleExecutor, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	cacheID := rm.genCacheID(path, method)
	execID, cached := rm.cache.Get(cacheID)
	if cached {
		return rm.executors[execID.(string)], true
	}

	for id, exec := range rm.executors {
		if exec.match(path, method) {
			rm.cache.Add(cacheID, id)
			return exec, true
		}
	}

	return nil, false
}

func (rm *ruleManager) createRule(rule *ResourceRule) (*ruleExecutor, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	re, err := newRuleExecutor(rule)
	if err != nil {
		return nil, err
	}
	_, ok := rm.executors[re.id()]
	if ok {
		Logger.Error("failed to create duplicated rule", zap.String("path", rule.Request.Path), zap.String("method", rule.Request.Method))
		return nil, errors.New("found duplicated rule")
	}
	rm.executors[re.id()] = re
	return re, nil
}

func init() {
	defaultRuleManager = newRuleManager()
}
