package domain

import (
	"bytes"
	"errors"
	"html/template"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"github.com/wosai/deepmock/misc"
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

var (
	defaultTemplateFuncs template.FuncMap
)

type (
	FilterMode = string

	Executor struct {
		ID          string
		Method      []byte
		Path        *regexp.Regexp
		Variable    map[string]interface{}
		Weight      WeightPicker
		Regulations []*RegulationExecutor
		Version     int
	}

	WeightPicker map[string]*WeightDice

	WeightDice struct {
		total        int
		distribution []string
		factor       map[string]uint
	}

	RegulationExecutor struct {
		IsDefault bool
		Filter    *FilterExecutor
		Template  *TemplateExecutor
	}

	TemplateExecutor struct {
		IsGolangTemplate bool
		IsBinData        bool
		template         *template.Template
		header           *fasthttp.ResponseHeader
		body             []byte
	}

	RenderContext struct {
		Variable map[string]interface{}
		Weight   map[string]string
		Header   map[string]string
		Query    map[string]string
		Form     map[string]string
		Json     map[string]interface{}
	}

	FilterExecutor struct {
		Query  *QueryFilterExecutor
		Header *HeaderFilterExecutor
		Body   *BodyFilterExecutor
	}

	BodyFilterExecutor struct {
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

func (wp WeightPicker) DiceAll() map[string]string {
	ret := make(map[string]string)
	for k, v := range wp {
		ret[k] = v.Dice()
	}
	return ret
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
		return hfe.filterByKeyword(header)

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
		return qfe.filterByKeyword(args)

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

func (te *TemplateExecutor) Render(ctx *fasthttp.RequestCtx, v map[string]interface{}, weight map[string]string) error {
	te.header.CopyTo(&ctx.Response.Header)
	if !te.IsGolangTemplate {
		ctx.Response.SetBody(te.body)
		return nil
	}

	// 开始渲染模板
	var rc RenderContext
	h := extractHeaderAsParams(&ctx.Request)
	q := extractQueryAsParams(&ctx.Request)
	f, j := extractBodyAsParams(&ctx.Request)

	rc.Variable = v
	rc.Weight = weight
	rc.Header = h
	rc.Query = q
	rc.Form = f
	rc.Json = j
	return te.template.Execute(ctx.Response.BodyWriter(), rc)
}

func (re *RegulationExecutor) Render(ctx *fasthttp.RequestCtx, v map[string]interface{}, w map[string]string) error {
	return re.Template.Render(ctx, v, w)
}

// todo:
func NewExecutor() (*Executor, error) {
	return nil, nil
}

func (exe *Executor) Match(path, method []byte) bool {
	if bytes.Compare(method, exe.Method) != 0 {
		return false
	}
	return exe.Path.Match(path)
}

func (exe *Executor) FindRegulationExecutor(request *fasthttp.Request) *RegulationExecutor {
	var reg *RegulationExecutor

	for _, regulation := range exe.Regulations {
		if regulation.IsDefault {
			reg = regulation
		}
		if regulation.Filter.Filter(request) {
			return regulation
		}
	}
	return reg
}

// RegisterTemplateFunc 注册模板自定义函数
func RegisterTemplateFunc(name string, f interface{}) error {
	if _, ok := defaultTemplateFuncs[name]; ok {
		return errors.New("func named " + name + " was exists")
	}
	defaultTemplateFuncs[name] = f
	return nil
}

func genUUID() string {
	return uuid.New().String()
}

func currentTimestamp(precision string) int64 {
	now := time.Now().UnixNano()
	switch precision {
	case "mcs":
		return now / 1e3
	case "ms":
		return now / 1e6
	case "sec":
		return now / 1e9
	default:
		return now
	}
}

func formatDate(layout string) string {
	return time.Now().Format(layout)
}

func plus(v interface{}, i int) interface{} {
	switch v.(type) {
	case int:
		return v.(int) + i
	case float64:
		return v.(float64) + float64(i)
	case float32:
		return v.(float32) + float32(i)
	case string:
		vv, err := strconv.Atoi(v.(string))
		if err != nil {
			return "unsupported type"
		}
		return vv + i
	default:
		return "unsupported type"
	}
}

func dateDelta(date, layout string, year, month, day int) string {
	t, err := time.Parse(layout, date)
	if err != nil {
		return date
	}
	return t.AddDate(year, month, day).Format(layout)
}

func init() {
	// create build-in template functions
	defaultTemplateFuncs = make(template.FuncMap)
	_ = RegisterTemplateFunc("uuid", genUUID)
	_ = RegisterTemplateFunc("timestamp", currentTimestamp)
	_ = RegisterTemplateFunc("date", formatDate)
	_ = RegisterTemplateFunc("plus", plus)
	_ = RegisterTemplateFunc("rand_string", misc.GenRandomString)
	_ = RegisterTemplateFunc("date_delta", dateDelta)
}
