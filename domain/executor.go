package domain

import (
	"bytes"
	"math/rand"
	"regexp"

	"github.com/valyala/fasthttp"
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

	ModeField = "mode"
)

type (
	FilterMode = string

	Executor struct {
		ID          string
		Method      []byte
		Path        *regexp.Regexp
		Variable    map[string]interface{}
		Weight      map[string]*WeightDice
		Regulations []*RegulationExecutor
		Version     int
	}

	WeightDice struct {
		total        int
		distribution []string
		factor       map[string]uint
	}

	RegulationExecutor struct {
		IsDefault bool
		Filter    *FilterExecutor
	}

	FilterExecutor struct {
		Query  *QueryFilterExecutor
		Header *HeaderFilterExecutor
		Body   *BodyFilterExecutor
	}

	BodyFilterExecutor struct {
		params  map[string][]byte
		mode    FilterMode
		regular *regexp.Regexp
		keyword []byte
	}

	HeaderFilterExecutor struct {
		params   map[string][]byte
		mode     FilterMode
		regulars map[string]*regexp.Regexp
	}

	QueryFilterExecutor struct {
		params   map[string][]byte
		mode     FilterMode
		regulars map[string]*regexp.Regexp
	}
)

func NewWeightDice(factor map[string]uint) *WeightDice {
	wd := &WeightDice{
		total:        0,
		distribution: []string{},
		factor:       factor,
	}

	for k, v := range factor {
		for i := 0; i < int(v); i++ {
			wd.distribution = append(wd.distribution, k)
			wd.total++
		}
	}
	return wd
}

func (wd *WeightDice) Dice() string {
	return wd.distribution[rand.Intn(wd.total)]
}

func (hfe *HeaderFilterExecutor) filterByExactKeyValue(header *fasthttp.RequestHeader) bool {
	for k, v := range hfe.params {
		if bytes.Compare(header.Peek(k), v) != 0 {
			return false
		}
	}
	return true
}

func (hfe *HeaderFilterExecutor) filterByKeyword(header *fasthttp.RequestHeader) bool {
	for k, v := range hfe.params {
		if !bytes.Contains(header.Peek(k), v) {
			return false
		}
	}
	return true
}

func (hfe *HeaderFilterExecutor) filterByRegular(header *fasthttp.RequestHeader) bool {
	for k := range hfe.params {
		if !hfe.regulars[k].Match(header.Peek(k)) {
			return false
		}
	}
	return true
}

func (hfe *HeaderFilterExecutor) Filter(header *fasthttp.RequestHeader) bool {
	if hfe == nil {
		return true
	}

	switch hfe.mode {
	case FilterModeAlwaysTrue:
		return true

	case FilterModeExact:
		return hfe.filterByExactKeyValue(header)

	case FilterModeKeyword:
		return hfe.filterByExactKeyValue(header)

	case FilterModeRegular:
		return hfe.filterByRegular(header)

	default:
		return false
	}
}

func (qfe *QueryFilterExecutor) filterByExactKeyValue(args *fasthttp.Args) bool {
	for k, v := range qfe.params {
		if bytes.Compare(args.Peek(k), v) != 0 {
			return false
		}
	}
	return true
}

func (qfe *QueryFilterExecutor) filterByKeyword(args *fasthttp.Args) bool {
	for k, v := range qfe.params {
		if !bytes.Contains(args.Peek(k), v) {
			return false
		}
	}
	return true
}

func (qfe *QueryFilterExecutor) filterByRegular(args *fasthttp.Args) bool {
	for k := range qfe.params {
		if !qfe.regulars[k].Match(args.Peek(k)) {
			return false
		}
	}
	return true
}

func (qfe *QueryFilterExecutor) Filter(args *fasthttp.Args) bool {
	if qfe == nil {
		return true
	}

	switch qfe.mode {
	case FilterModeAlwaysTrue:
		return true

	case FilterModeExact:
		return qfe.filterByExactKeyValue(args)

	case FilterModeKeyword:
		return qfe.filterByExactKeyValue(args)

	case FilterModeRegular:
		return qfe.filterByRegular(args)

	default:
		return false
	}
}

func (bfe *BodyFilterExecutor) Filter(body []byte) bool {
	if bfe == nil {
		return true
	}

	switch bfe.mode {
	case FilterModeAlwaysTrue:
		return true

	case FilterModeKeyword:
		return bytes.Contains(body, bfe.keyword)

	case FilterModeRegular:
		return bfe.regular.Match(body)

	default:
		return false
	}
}

func (fe *FilterExecutor) Filter(request *fasthttp.Request) bool {
	if fe == nil {
		return true
	}
	if !fe.Header.Filter(&request.Header) {
		return false
	}
	if !fe.Query.Filter(request.URI().QueryArgs()) {
		return false
	}
	if !fe.Body.Filter(request.Body()) {
		return false
	}

	return true
}

func NewExecutor() (*Executor, error) {
	return nil, nil
}

func (exe *Executor) Match(path, method []byte) bool {
	if bytes.Compare(method, exe.Method) != 0 {
		return false
	}
	return exe.Path.Match(path)
}
