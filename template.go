package deepmock

import (
	"encoding/base64"
	"errors"
	"html/template"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/qastub/deepmock/types"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type (
	// responseTemplate http响应报文模板
	responseTemplate struct {
		isTemplate   bool
		isBinData    bool
		header       *fasthttp.ResponseHeader
		body         []byte
		htmlTemplate *template.Template
		raw          *types.ResourceResponseTemplate
	}

	renderContext struct {
		Variable ruleVariable
		Weight   params
		Header   params
		Query    params
		Form     params
		Json     map[string]interface{}
	}

	params map[string]string
)

var (
	defaultTemplateFuncs template.FuncMap
)

func newResponseTemplate(rrt *types.ResourceResponseTemplate) (*responseTemplate, error) {
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
		tmpl, err := template.New(genRandomString(8)).Funcs(defaultTemplateFuncs).Parse(string(rt.body))
		if err != nil {
			Logger.Error("failed to parse html template", zap.ByteString("template", rt.body), zap.Error(err))
			return nil, err
		}
		rt.htmlTemplate = tmpl
	}
	return rt, nil
}

func (rt *responseTemplate) renderTemplate(rc renderContext, w io.Writer) error {
	if !rt.isTemplate {
		return nil
	}
	return rt.htmlTemplate.Execute(w, rc)
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

// RegisterTemplateFunc 注册模板自定义函数
func RegisterTemplateFunc(name string, f interface{}) error {
	if _, ok := defaultTemplateFuncs[name]; ok {
		return errors.New("func named " + name + " was exists")
	}
	defaultTemplateFuncs[name] = f
	return nil
}

func init() {
	// create build-in template functions
	defaultTemplateFuncs = make(template.FuncMap)
	_ = RegisterTemplateFunc("uuid", genUUID)
	_ = RegisterTemplateFunc("timestamp", currentTimestamp)
	_ = RegisterTemplateFunc("date", formatDate)
	_ = RegisterTemplateFunc("plus", plus)
	_ = RegisterTemplateFunc("rand_string", genRandomString)
	_ = RegisterTemplateFunc("date_delta", dateDelta)
}
