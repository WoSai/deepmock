package infrastructure

import (
	"bytes"
	"context"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/wosai/deepmock/domain"
	"github.com/wosai/deepmock/misc"
	"go.uber.org/zap"
)

type (
	ExecutorRepository struct {
		executors map[string]*domain.Executor
		cache     *lru.ARCCache
		mu        sync.RWMutex
	}
)

var (
	delimiter = []byte("-")
)

func NewExecutorRepository(size int) *ExecutorRepository {
	cache, err := lru.NewARC(size)
	if err != nil {
		panic(err)
	}

	return &ExecutorRepository{
		executors: map[string]*domain.Executor{},
		cache:     cache,
	}
}

func (er *ExecutorRepository) cacheID(path, method []byte) []byte {
	return bytes.Join([][]byte{path, method}, delimiter)
}

func (er *ExecutorRepository) FindExecutor(_ context.Context, path, method []byte) (*domain.Executor, bool) {
	cid := er.cacheID(path, method)
	val, cached := er.cache.Get(cid)
	// 如果存在缓存，需要再次从executors确认是否还在
	if cached {
		er.mu.RLock()
		exe, exists := er.executors[val.(string)]
		er.mu.RUnlock()

		if exists {
			return exe, true
		}
		er.cache.Remove(cid) // 已经失效
		return nil, false
	}

	// 不存在时，需要用正则匹配规则
	er.mu.RLock()
	for eid, executor := range er.executors {
		if executor.Match(path, method) {
			er.mu.RUnlock()
			er.cache.Add(cid, eid)
			return executor, true
		}
	}
	er.mu.RUnlock()
	return nil, false
}

func (er *ExecutorRepository) Purge(_ context.Context) {
	er.mu.Lock()
	defer er.mu.Unlock()

	for k := range er.executors {
		delete(er.executors, k)
	}
	er.cache.Purge()
}

func (er *ExecutorRepository) ImportAll(_ context.Context, executors ...*domain.Executor) {
	er.mu.Lock()
	defer er.mu.Unlock()

	toDelete := make(map[string]struct{}, len(er.executors))
	for k := range er.executors {
		toDelete[k] = struct{}{}
	}

	for _, executor := range executors {
		current, exists := er.executors[executor.ID]
		delete(toDelete, executor.ID)
		if exists && current.Version == executor.Version { // 记录未变更
			continue
		}
		er.executors[executor.ID] = executor // 记录不存在或者版本不同了，都变更
	}

	// toDelete中如果还存在数据，即表示需要删除
	if len(toDelete) > 0 {

		for k, _ := range toDelete {
			misc.Logger.Info("deleted expired rules", zap.String("rule_id", k))
			delete(er.executors, k)
		}
	}
}
