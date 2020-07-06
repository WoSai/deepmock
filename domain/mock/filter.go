package mock

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/wosai/deepmock/misc"

	"github.com/valyala/fasthttp"
	"github.com/wosai/deepmock/types/resource"
	"go.uber.org/zap"
)

type (
	// headerFilter 请求头筛选器，如果为空，默认通过
	headerFilter struct {
		params   resource.HeaderFilterParameters
		mode     resource.FilterMode
		regulars map[string]*regexp.Regexp
	}

	// bodyFilter 请求body部分筛选器，如果为空，默认通过
	bodyFilter struct {
		params  resource.BodyFilterParameters
		mode    resource.FilterMode
		regular *regexp.Regexp
		keyword []byte
	}

	// queryFilter 请求query string筛选器，如果为空，默认通过
	queryFilter struct {
		params   resource.QueryFilterParameters
		mode     resource.FilterMode
		regulars map[string]*regexp.Regexp
	}

	// requestFilter http请求筛选器
	requestFilter struct {
		header *headerFilter
		query  *queryFilter
		body   *bodyFilter
	}
)

func (hf *headerFilter) withParameters(p resource.HeaderFilterParameters) error {
	hf.params = p
	if hf.params != nil {
		hf.mode = hf.params["mode"]
		delete(hf.params, "mode")
	} else { // 必须这么判断，否则存在mode值不刷新的可能
		hf.mode = resource.FilterModeAlwaysTrue
	}

	if hf.mode == resource.FilterModeRegular {
		hf.regulars = map[string]*regexp.Regexp{}

		var err error
		for k, v := range hf.params {
			hf.regulars[k], err = regexp.Compile(v)
			if err != nil {
				misc.Logger.Error("failed to compile regular expression", zap.String("expression", v), zap.Error(err))
				return err
			}
		}
	}

	return nil
}

func (hf *headerFilter) wrap() resource.HeaderFilterParameters {
	if hf.params == nil {
		return nil
	}
	ret := make(resource.HeaderFilterParameters)
	for k, v := range hf.params {
		ret[k] = v
	}
	ret["mode"] = hf.mode
	return ret
}

func (hf *headerFilter) filter(h *fasthttp.RequestHeader) bool {
	if hf == nil {
		return true
	}

	switch hf.mode {
	case resource.FilterModeAlwaysTrue:
		return true

	case resource.FilterModeExact:
		return hf.filterByExactKeyValue(h)

	case resource.FilterModeKeyword:
		return hf.filterByKeyword(h)

	case resource.FilterModeRegular:
		return hf.filterByRegular(h)

	default:
		misc.Logger.Warn("found unsupported filter mode in headerFilter", zap.String("mode", hf.mode))
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

func (bf *bodyFilter) withParameters(params resource.BodyFilterParameters) error {
	bf.params = params
	if bf.params != nil {
		bf.mode = bf.params["mode"]
		delete(bf.params, "mode")
	} else {
		bf.mode = resource.FilterModeAlwaysTrue
	}

	switch bf.mode {
	case resource.FilterModeRegular:
		var err error
		bf.regular, err = regexp.Compile(bf.params["regular"])
		if err != nil {
			misc.Logger.Error("failed to compile regular expression", zap.String("expression", bf.params["regular"]), zap.Error(err))
			return err
		}
		delete(bf.params, "regular")

	case resource.FilterModeKeyword:
		bf.keyword = []byte(bf.params["keyword"])
		delete(bf.params, "keyword")
	}
	return nil
}

func (bf *bodyFilter) wrap() resource.BodyFilterParameters {
	if bf.params == nil {
		return nil
	}
	ret := make(resource.BodyFilterParameters)
	for k, v := range bf.params {
		ret[k] = v
	}
	switch bf.mode {
	case resource.FilterModeRegular:
		ret["regular"] = bf.regular.String()
		fallthrough
	case resource.FilterModeKeyword:
		ret["keyword"] = string(bf.keyword)
		fallthrough
	default:
		ret["mode"] = bf.mode
	}
	return ret
}

func (bf *bodyFilter) filter(body []byte) bool {
	if bf == nil {
		return true
	}

	switch bf.mode {
	case resource.FilterModeAlwaysTrue:
		return true

	case resource.FilterModeKeyword:
		return bytes.Contains(body, bf.keyword)

	case resource.FilterModeRegular:
		return bf.regular.Match(body)

	default:
		misc.Logger.Warn("found unsupported filter mode in bodyFilter", zap.String("mode", bf.mode))
		return false
	}
}

func (qf *queryFilter) withParameters(query resource.QueryFilterParameters) error {
	qf.params = query
	if qf.params != nil {
		qf.mode = qf.params["mode"]
		delete(qf.params, "mode")
	} else {
		qf.mode = resource.FilterModeAlwaysTrue
	}

	if qf.mode == resource.FilterModeRegular {
		var err error
		qf.regulars = map[string]*regexp.Regexp{}
		for k, v := range qf.params {
			qf.regulars[k], err = regexp.Compile(v)
			if err != nil {
				misc.Logger.Error("failed to compile regular expression", zap.String("expression", v), zap.Error(err))
				return err
			}
		}
	}
	return nil
}

func (qf *queryFilter) wrap() resource.QueryFilterParameters {
	if qf.params == nil {
		return nil
	}
	ret := make(resource.QueryFilterParameters)
	for k, v := range qf.params {
		ret[k] = v
	}
	ret["mode"] = qf.mode
	return ret
}

func (qf *queryFilter) filter(query *fasthttp.Args) bool {
	if qf == nil {
		return true
	}

	switch qf.mode {
	case resource.FilterModeAlwaysTrue:
		return true

	case resource.FilterModeExact:
		return qf.filterByExactKeyValue(query)

	case resource.FilterModeKeyword:
		return qf.filterByKeyword(query)

	case resource.FilterModeRegular:
		return qf.filterByRegular(query)

	default:
		misc.Logger.Warn("found unsupported filter mode in queryFilter", zap.String("mode", qf.mode))
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

func newRequestFilter(f *resource.Filter) (*requestFilter, error) {
	if f == nil {
		return nil, nil
	}
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

func (rf *requestFilter) wrap() *resource.Filter {
	if rf == nil {
		return nil
	}
	f := new(resource.Filter)
	f.Header = rf.header.wrap()
	f.Query = rf.query.wrap()
	f.Body = rf.body.wrap()
	return f
}

func (rf *requestFilter) filter(req *fasthttp.Request) bool {
	if rf == nil {
		return true
	}
	if !rf.header.filter(&req.Header) {
		return false
	}

	if !rf.query.filter(req.URI().QueryArgs()) {
		return false
	}

	if !rf.body.filter(req.Body()) {
		return false
	}
	return true
}
