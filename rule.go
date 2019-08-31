package deepmock

import (
	"bytes"
	"errors"
	"regexp"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/qastub/deepmock/types"
	"github.com/valyala/fasthttp"
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
		responseRegulations responseRegulationSet
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

	responseRegulationSet []*responseRegulation
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
	rule.Responses = make(types.ResourceResponseRegulationSet, len(re.responseRegulations))
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
	for i, regulation := range re.responseRegulations {
		if regulation.isDefault {
			d = regulation
			Logger.Info("found the default response regulation", zap.Int("index", i), zap.String("rule", re.id()))
		}
		if regulation.filter(req) {
			re.mu.RUnlock()
			Logger.Info("hit the response regulation", zap.Int("index", i), zap.String("rule", re.id()))
			return regulation
		}
	}
	re.mu.RUnlock()
	Logger.Info("no response regulation hit, use default response", zap.String("rule", re.id()))
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
	if cached { // 缓存中存在时，不代表该ruleExecutor一定存在，有可能规则已经删除，但未清理缓存
		re, exists := rm.executors[execID.(string)]
		if exists {
			return re, true
		}
		rm.cache.Remove(cacheID)
		return nil, false
	}

	for id, exec := range rm.executors {
		if exec.match(path, method) {
			rm.cache.Add(cacheID, id)
			return exec, true
		}
	}

	return nil, false
}

func (rm *ruleManager) updateRule(rule *types.ResourceRule) (*ruleExecutor, error) {
	re, err := newRuleExecutor(rule)
	if err != nil {
		return nil, err
	}

	rm.mu.Lock()
	_, ok := rm.executors[re.id()]
	if !ok {
		rm.mu.Unlock()
		Logger.Error("the rule to update is not exists", zap.String("path", rule.Request.Path), zap.String("method", rule.Request.Method))
		return nil, errors.New("the rule to update is not exists")
	}
	rm.executors[re.id()] = re
	rm.mu.Unlock()
	Logger.Info("rule is updated", zap.ByteString("path", re.requestMatcher.path), zap.ByteString("method", re.requestMatcher.method))
	return re, nil
}

func (rm *ruleManager) createRule(rule *types.ResourceRule) (*ruleExecutor, error) {
	rm.mu.Lock()
	re, err := rm.createRuleInto(rule, rm.executors)
	rm.mu.Unlock()
	return re, err
}

func (rm *ruleManager) createRuleInto(rule *types.ResourceRule, m map[string]*ruleExecutor) (*ruleExecutor, error) {
	re, err := newRuleExecutor(rule)
	if err != nil {
		return nil, err
	}
	_, ok := rm.executors[re.id()]
	if ok {
		Logger.Error("failed to create duplicated rule", zap.String("path", rule.Request.Path), zap.String("method", rule.Request.Method))
		return nil, errors.New("found duplicated rule")
	}
	m[re.id()] = re
	Logger.Info("created new rule", zap.ByteString("path", re.requestMatcher.path), zap.ByteString("method", re.requestMatcher.method))
	return re, nil
}

func (rm *ruleManager) batchCreateRules(rules ...*types.ResourceRule) error {
	rm.mu.Lock()
	for _, rule := range rules {
		if _, err := rm.createRuleInto(rule, rm.executors); err != nil {
			rm.mu.Unlock()
			return err
		}
	}
	rm.mu.Unlock()
	return nil
}

func (rm *ruleManager) deleteRule(res *types.ResourceRule) {
	rm.mu.Lock()
	delete(rm.executors, res.ID) // 不从缓存冲删除，因为无法获取cacheID
	rm.mu.Unlock()
	Logger.Info("delete rule with id " + res.ID)
}

func (rm *ruleManager) patchRule(res *types.ResourceRule) (*ruleExecutor, error) {
	rm.mu.RLock()
	re, exists := rm.executors[res.ID]
	rm.mu.RUnlock()

	if !exists {
		err := errors.New("cannot patch not exists rule")
		Logger.Error(err.Error())
		return nil, err
	}

	err := re.patch(res)
	if err == nil {
		Logger.Info("success patch rule", zap.ByteString("patch", re.requestMatcher.path), zap.ByteString("method", re.requestMatcher.method))
	}
	return re, err
}

func (rm *ruleManager) getRuleByID(i string) (*ruleExecutor, bool) {
	rm.mu.RLock()
	re, exists := rm.executors[i]
	rm.mu.RUnlock()
	return re, exists
}

func (rm *ruleManager) getRule(res *types.ResourceRule) (*ruleExecutor, bool) {
	rm.mu.RLock()
	re, exists := rm.executors[res.ID]
	rm.mu.RUnlock()
	return re, exists
}

func (rm *ruleManager) exportRules() []*ruleExecutor {
	rm.mu.RLock()
	es := make([]*ruleExecutor, len(rm.executors))
	var count int
	for _, v := range rm.executors {
		es[count] = v
		count++
	}
	rm.mu.RUnlock()
	return es
}

func (rm *ruleManager) importRules(rules ...*types.ResourceRule) error {
	ne := make(map[string]*ruleExecutor)
	for _, rule := range rules {
		if _, err := rm.createRuleInto(rule, ne); err != nil {
			return err
		}
	}

	rm.mu.Lock()
	rm.executors = ne
	rm.cache.Purge()
	rm.mu.Unlock()
	Logger.Info("success import rules")
	return nil
}

func (rm *ruleManager) reset() {
	rm.mu.Lock()

	for k := range rm.executors {
		delete(rm.executors, k)
	}

	rm.cache.Purge()

	rm.mu.Unlock()
}

func init() {
	defaultRuleManager = newRuleManager()
}
