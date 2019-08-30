package deepmock

import (
	"bytes"
	"errors"
	"regexp"
	"sync"

	"github.com/valyala/fasthttp"

	lru "github.com/hashicorp/golang-lru"
	"github.com/qastub/deepmock/types"
	"go.uber.org/zap"
)

type (
	ruleContext types.ResourceContext

	ruleManager struct {
		executors map[string]*ruleExecutor
		cache     *lru.ARCCache
		mu        sync.RWMutex
	}

	ruleExecutor struct {
		requestMatcher      *requestMatcher
		context             ruleContext
		weightPicker        weightingPicker
		responseRegulations []*responseRegulation
		mu                  sync.RWMutex
	}

	// requestMatcher 请求匹配器
	requestMatcher struct {
		id     string
		path   []byte
		method []byte
		re     *regexp.Regexp
		raw    *types.ResourceRequestMatcher
	}
)

var (
	defaultRuleManager *ruleManager
)

func newRequestMatcher(res *types.ResourceRequestMatcher) (*requestMatcher, error) {
	if err := res.Check(); err != nil {
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

func (rm *requestMatcher) wrap() *types.ResourceRequestMatcher {
	return rm.raw
}

func (rc ruleContext) patch(res types.ResourceContext) {
	for k, v := range res {
		rc[k] = v
	}
}

func newRuleExecutor(res *types.ResourceRule) (*ruleExecutor, error) {
	re := new(ruleExecutor)
	// init requestMatcher
	rm, err := newRequestMatcher(res.Request)
	if err != nil {
		return nil, err
	}
	re.requestMatcher = rm
	rm.raw = res.Request

	// init context
	re.context = ruleContext(res.Context)

	// init weight
	re.weightPicker = map[string]*weightingDice{}
	for k, v := range res.Weight {
		re.weightPicker[k] = newWeighingDice(v)
	}

	// init responseRegulation
	if err := res.Responses.Check(); err != nil {
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

func (re *ruleExecutor) wrap() *types.ResourceRule {
	re.mu.RLock()

	rule := new(types.ResourceRule)
	rule.ID = re.id()
	rule.Request = re.requestMatcher.wrap()
	rule.Context = types.ResourceContext(re.context)
	rule.Weight = re.weightPicker.wrap()
	rule.Responses = types.ResourceResponseRegulationSet{}
	for k, v := range re.responseRegulations {
		rule.Responses[k] = v.wrap()
	}

	re.mu.RUnlock()
	return rule
}

func (re *ruleExecutor) patch(res *types.ResourceRule) error {
	re.mu.Lock()

	if res.Responses != nil {
		if err := res.Responses.Check(); err != nil {
			return err
		}
		rs := make([]*responseRegulation, len(res.Responses))
		for k, v := range res.Responses {
			rr, err := newResponseRegulation(v)
			if err != nil {
				return err
			}
			rs[k] = rr
		}
		re.responseRegulations = rs
	}

	if res.Context != nil {
		re.context.patch(res.Context)
	}

	if res.Weight != nil {
		re.weightPicker.patch(res.Weight)
	}

	re.mu.Unlock()
	return nil
}

func (re *ruleExecutor) visitBy(req *fasthttp.Request) *responseRegulation {
	re.mu.RLock()

	var d *responseRegulation
	for _, regulation := range re.responseRegulations {
		if regulation.isDefault {
			d = regulation
		}
		if regulation.filter(req) {
			re.mu.RUnlock()
			return regulation
		}
	}
	re.mu.RUnlock()
	return d
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

func (rm *ruleManager) createRule(rule *types.ResourceRule) (*ruleExecutor, error) {
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
