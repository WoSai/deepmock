package deepmock

import (
	"bytes"
	"regexp"
	"strings"
	"sync"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type (
	FilterMode = string

	headerFilter struct {
		params   HeaderFilterParameters
		mode     FilterMode
		regulars map[string]*regexp.Regexp
		mu       sync.RWMutex
	}

	bodyFilter struct {
		params  BodyFilterParameters
		mode    FilterMode
		regular *regexp.Regexp
		keyword []byte
		mu      sync.RWMutex
	}

	queryFilter struct {
		params        QueryFilterParameters
		mode          FilterMode
		regulars      map[string]*regexp.Regexp
		escapedValues map[string]string
		mu            sync.RWMutex
	}
)

const (
	FilterModeAlwaysTrue FilterMode = "always_true"
	FilterModeExact      FilterMode = "exact"
	FilterModeKeyword    FilterMode = "keyword"
	FilterModeRegular    FilterMode = "regular"
)

func (hf *headerFilter) withParameters(p HeaderFilterParameters) error {
	hf.mu.Lock()
	defer hf.mu.Unlock()

	hf.params = p
	if hf.params != nil {
		hf.mode = hf.params["mode"]
		if hf.mode == "" {
			hf.mode = FilterModeAlwaysTrue
		}
		delete(hf.params, "mode")
	} else { // 必须这么判断，否则存在mode值不刷新的可能
		hf.mode = FilterModeAlwaysTrue
	}

	if hf.mode == FilterModeRegular {
		hf.regulars = map[string]*regexp.Regexp{}

		var err error
		for k, v := range hf.params {
			hf.regulars[k], err = regexp.Compile(v)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (hf *headerFilter) filter(h *fasthttp.RequestHeader) bool {
	hf.mu.RLock()
	defer hf.mu.RUnlock()

	switch hf.mode {
	case FilterModeAlwaysTrue:
		return true

	case FilterModeExact:
		return hf.filterByExactKeyValue(h)

	case FilterModeKeyword:
		return hf.filterByKeyword(h)

	case FilterModeRegular:
		return hf.filterByRegular(h)

	default:
		Logger.Warn("found unsupported filter mode in headerFilter", zap.String("mode", hf.mode))
		return false
	}
}

func (hf *headerFilter) filterByExactKeyValue(h *fasthttp.RequestHeader) bool {
	for k, v := range hf.params {
		if string(h.Peek(k)) != v {
			return false
		}
	}
	return true
}

func (hf *headerFilter) filterByKeyword(h *fasthttp.RequestHeader) bool {
	for k, v := range hf.params {
		actual := string(h.Peek(k))
		if !strings.Contains(actual, v) {
			return false
		}
	}
	return true
}

func (hf *headerFilter) filterByRegular(h *fasthttp.RequestHeader) bool {
	for k := range hf.params {
		if !hf.regulars[k].Match(h.Peek(k)) {
			return false
		}
	}
	return true
}

func (bf *bodyFilter) withParameters(params BodyFilterParameters) error {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	bf.params = params
	if bf.params != nil {
		bf.mode = bf.params["mode"]
		if bf.mode == "" {
			bf.mode = FilterModeAlwaysTrue
		}
		delete(bf.params, "mode")
	} else {
		bf.mode = FilterModeAlwaysTrue
	}

	switch bf.mode {
	case FilterModeRegular:
		var err error
		bf.regular, err = regexp.Compile(bf.params["regular"])
		if err != nil {
			return err
		}
		delete(bf.params, "regular")

	case FilterModeKeyword:
		bf.keyword = []byte(bf.params["keyword"])
		delete(bf.params, "keyword")
	}
	return nil
}

func (bf *bodyFilter) filter(body []byte) bool {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	switch bf.mode {
	case FilterModeAlwaysTrue:
		return true

	case FilterModeKeyword:
		return bytes.Contains(body, bf.keyword)

	case FilterModeRegular:
		return bf.regular.Match(body)

	default:
		Logger.Warn("found unsupported filter mode in bodyFilter", zap.String("mode", bf.mode))
		return false
	}
}

func (qf *queryFilter) withParameters(query QueryFilterParameters) error {
	qf.mu.Lock()
	defer qf.mu.Unlock()

	qf.params = query
	if qf.params != nil {
		qf.mode = qf.params["mode"]
		if qf.mode == "" {
			qf.mode = FilterModeAlwaysTrue
		}
		delete(qf.params, "mode")
	} else {
		qf.mode = FilterModeAlwaysTrue
	}

	if qf.mode == FilterModeRegular {
		var err error
		qf.regulars = map[string]*regexp.Regexp{}
		for k, v := range qf.params {
			qf.regulars[k], err = regexp.Compile(v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (qf *queryFilter) filter(query *fasthttp.Args) bool {
	qf.mu.RLock()
	defer qf.mu.RUnlock()

	switch qf.mode {
	case FilterModeAlwaysTrue:
		return true

	case FilterModeExact:
		return qf.filterByExactKeyValue(query)

	case FilterModeKeyword:
		return qf.filterByKeyword(query)

	case FilterModeRegular:
		return qf.filterByRegular(query)

	default:
		Logger.Warn("found unsupported filter mode in queryFilter", zap.String("mode", qf.mode))
		return false
	}
}

func (qf *queryFilter) filterByExactKeyValue(query *fasthttp.Args) bool {
	for k, v := range qf.params {
		if v != string(query.Peek(k)) {
			return false
		}
	}
	return true
}

func (qf *queryFilter) filterByKeyword(query *fasthttp.Args) bool {
	for k, v := range qf.params {
		if !bytes.Contains(query.Peek(k), []byte(v)) {
			return false
		}
	}
	return true
}

func (qf *queryFilter) filterByRegular(query *fasthttp.Args) bool {
	for k := range qf.regulars {
		if !qf.regulars[k].Match(query.Peek(k)) {
			return false
		}
	}
	return true
}
