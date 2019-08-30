package deepmock

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/qastub/deepmock/types"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type (
	// FilterMode 筛选模式
	FilterMode = string

	// headerFilter 请求头筛选器，如果为空，默认通过
	headerFilter struct {
		params   types.ResourceHeaderFilterParameters
		mode     FilterMode
		regulars map[string]*regexp.Regexp
	}

	// bodyFilter 请求body部分筛选器，如果为空，默认通过
	bodyFilter struct {
		params  types.ResourceBodyFilterParameters
		mode    FilterMode
		regular *regexp.Regexp
		keyword []byte
	}

	// queryFilter 请求query string筛选器，如果为空，默认通过
	queryFilter struct {
		params   types.ResourceQueryFilterParameters
		mode     FilterMode
		regulars map[string]*regexp.Regexp
	}

	// requestFilter http请求筛选器
	requestFilter struct {
		header *headerFilter
		query  *queryFilter
		body   *bodyFilter
	}
)

const (
	// FilterModeAlwaysTrue 总是通过
	FilterModeAlwaysTrue FilterMode = "always_true"
	// FilterModeExact key/value精确模式
	FilterModeExact FilterMode = "exact"
	// FilterModeKeyword 关键字模板，即确保contains(a, b)结果为true
	FilterModeKeyword FilterMode = "keyword"
	// FilterModeRegular 正则表达式模式
	FilterModeRegular FilterMode = "regular"
)

func (hf *headerFilter) withParameters(p types.ResourceHeaderFilterParameters) error {
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
				Logger.Error("failed to compile regular expression", zap.String("expression", v), zap.Error(err))
				return err
			}
		}
	}

	return nil
}

func (hf *headerFilter) wrap() types.ResourceHeaderFilterParameters {
	if hf.params == nil {
		return nil
	}
	ret := make(types.ResourceHeaderFilterParameters)
	for k, v := range hf.params {
		ret[k] = v
	}
	ret["mode"] = hf.mode
	return ret
}

func (hf *headerFilter) filter(h *fasthttp.RequestHeader) bool {
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

func (bf *bodyFilter) withParameters(params types.ResourceBodyFilterParameters) error {
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
			Logger.Error("failed to compile regular expression", zap.String("expression", bf.params["regular"]), zap.Error(err))
			return err
		}
		delete(bf.params, "regular")

	case FilterModeKeyword:
		bf.keyword = []byte(bf.params["keyword"])
		delete(bf.params, "keyword")
	}
	return nil
}

func (bf *bodyFilter) wrap() types.ResourceBodyFilterParameters {
	if bf.params == nil {
		return nil
	}
	ret := make(types.ResourceBodyFilterParameters)
	for k, v := range bf.params {
		ret[k] = v
	}
	switch bf.mode {
	case FilterModeRegular:
		ret["regular"] = bf.regular.String()
		fallthrough
	case FilterModeKeyword:
		ret["keyword"] = string(bf.keyword)
		fallthrough
	default:
		ret["mode"] = bf.mode
	}
	return ret
}

func (bf *bodyFilter) filter(body []byte) bool {
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

func (qf *queryFilter) withParameters(query types.ResourceQueryFilterParameters) error {
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
				Logger.Error("failed to compile regular expression", zap.String("expression", v), zap.Error(err))
				return err
			}
		}
	}
	return nil
}

func (qf *queryFilter) wrap() types.ResourceQueryFilterParameters {
	if qf.params == nil {
		return nil
	}
	ret := make(types.ResourceQueryFilterParameters)
	for k, v := range qf.params {
		ret[k] = v
	}
	ret["mode"] = qf.mode
	return ret
}

func (qf *queryFilter) filter(query *fasthttp.Args) bool {
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

func newRequestFilter(f *types.ResourceFilter) (*requestFilter, error) {
	var err error
	h := new(headerFilter)
	if err = h.withParameters(f.Header); err != nil {
		return nil, err
	}

	q := new(queryFilter)
	if err = q.withParameters(f.Query); err != nil {
		return nil, err
	}

	b := new(bodyFilter)
	if err = b.withParameters(f.Body); err != nil {
		return nil, err
	}

	return &requestFilter{header: h, query: q, body: b}, nil
}

func (rf *requestFilter) wrap() *types.ResourceFilter {
	f := new(types.ResourceFilter)
	f.Header = rf.header.wrap()
	f.Query = rf.query.wrap()
	f.Body = rf.body.wrap()
	return f
}

func (rf *requestFilter) filter(req *fasthttp.Request) bool {
	if !rf.header.filter(&req.Header) {
		return false
	}

	if !rf.query.filter(req.URI().QueryArgs()) {
		return false
	}

	if rf.body.filter(req.Body()) {
		return false
	}
	return true
}
