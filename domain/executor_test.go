package domain

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"html/template"
	"net/url"
	"regexp"
	"testing"
)

func TestHeaderFilter_Filter(t *testing.T) {
	var hp HeaderFilterParams

	hf, err := hp.To()
	assert.NoError(t, err)
	assert.Equal(t, hf.mode, FilterModeAlwaysTrue)
	assert.True(t, hf.Filter(nil))

	hf, err = HeaderFilterParams{"content-type": "application/json", "mode": "exact"}.To()
	assert.NoError(t, err)
	header := new(fasthttp.RequestHeader)
	header.SetContentType("application/json")
	assert.True(t, hf.Filter(header))
	header.SetContentType("application/json; charset=utf-8")
	assert.False(t, hf.Filter(header))

	hf, err = HeaderFilterParams{"content-type": "application/json", "x-trace-id": "foobar", "mode": "exact"}.To()
	assert.NoError(t, err)
	header = new(fasthttp.RequestHeader)
	header.SetContentType("application/json")
	assert.False(t, hf.Filter(header))
	header.SetContentType("application/json; charset=utf-8")
	assert.False(t, hf.Filter(header))

	hf, err = HeaderFilterParams{"content-type": "application/json", "mode": "keyword"}.To()
	assert.NoError(t, err)
	header = new(fasthttp.RequestHeader)
	header.SetContentType("application/json; charset=utf-8")
	assert.True(t, hf.Filter(header))
	header.SetContentType("application/xml")
	assert.False(t, hf.Filter(header))

	hf, err = HeaderFilterParams{"authCode": "[0-9]+", "mode": "regular"}.To()
	assert.NoError(t, err)
	header = new(fasthttp.RequestHeader)
	assert.False(t, hf.Filter(header))
	header.Set("authCode", "123123")
	assert.True(t, hf.Filter(header))
	header.Set("authCode", "hello world")
	assert.False(t, hf.Filter(header))
}

func TestBodyFilter_Filter(t *testing.T) {
	var params BodyFilterParams
	bf, err := params.To()
	assert.NoError(t, err)
	assert.True(t, bf.Filter(nil))

	bf, err = BodyFilterParams{"keyword": "foobar", "mode": "keyword"}.To()
	assert.NoError(t, err)
	assert.True(t, bf.Filter([]byte(`hello foobar`)))
	assert.False(t, bf.Filter([]byte(`hello world`)))

	bf, err = BodyFilterParams{"regular": "[0-9]+", "mode": "regular"}.To()
	assert.NoError(t, err)
	assert.False(t, bf.Filter([]byte(`what's your mobile phone number'`)))
	assert.True(t, bf.Filter([]byte(`my phone number is 110`)))
}

func TestQueryFilter_Filter(t *testing.T) {
	assertion := assert.New(t)

	qf, err := QueryFilterParams{}.To()
	assertion.NoError(err)
	assertion.True(qf.Filter(nil))

	qf, err = QueryFilterParams{"nation": "中国", "mode": "exact"}.To()
	assertion.NoError(err)
	query := new(fasthttp.Args)
	query.Set("version", "1")
	query.Set("nation", "中国")
	assertion.True(qf.Filter(query))
	query.Set("nation", "USA")
	assertion.False(qf.Filter(query))
	query.Del("nation")
	assertion.False(qf.Filter(query))

	qf, err = QueryFilterParams{"nation": "中国", "mode": "keyword"}.To()
	assertion.NoError(err)
	query = new(fasthttp.Args)
	query.Set("nation", "CHINA")
	assertion.False(qf.Filter(query))
	query.Set("nation", "中国123")
	assertion.True(qf.Filter(query))

	qf, err = QueryFilterParams{"age": "[0-9]+", "mode": "regular"}.To()
	assertion.NoError(err)
	query = new(fasthttp.Args)
	query.Set("age", "18")
	assertion.True(qf.Filter(query))
	query.Set("age", "unknown")
	assertion.False(qf.Filter(query))
}

func TestEmptyFilterExecutor(t *testing.T) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("POST")
	req.SetRequestURI("/api/v1/query")
	req.URI().SetQueryString("start=2019-09-01&end=2019-09-02")
	req.Header.SetContentType("application/json; charset=UTF-8")
	req.Header.Set("X-Version", "1.0")
	req.SetBody([]byte(`{"hello":"deepmock"}`))

	fe := new(FilterExecutor)
	assert.True(t, fe.Filter(req))

	fe = &FilterExecutor{}
	assert.True(t, fe.Filter(req))
}

func TestNewResponseTemplate(t *testing.T) {
	res := &Template{
		IsTemplate:     true,
		Header:         map[string]string{"Content-Type": "application/json", "Authorization": "123123"},
		StatusCode:     500,
		Body:           "hello world",
		B64EncodedBody: "aGVsbG8gZm9vYmFyIQ==",
	}

	executor, err := res.To()
	assert.NoError(t, err)
	assert.True(t, executor.IsGolangTemplate)
	assert.True(t, executor.IsBinData)
	assert.Equal(t, executor.body, []byte("hello foobar!"))
	assert.NotNil(t, executor.template)
	assert.Equal(t, executor.header.StatusCode(), 500)
	assert.Equal(t, executor.header.ContentType(), []byte("application/json"))
	assert.Equal(t, executor.header.Peek("Authorization"), []byte("123123"))
}

func TestUUIDFunc(t *testing.T) {
	text := `{{uuid}}`
	tmpl, err := template.New("test").Funcs(defaultTemplateFuncs).Parse(text)
	assert.Nil(t, err)

	buff := bytes.NewBuffer(nil)
	ctx := RenderContext{}
	assert.Nil(t, tmpl.Execute(buff, ctx))
	assert.Equal(t, len(buff.String()), 36)
}

func TestDateFunc(t *testing.T) {
	tmpl, err := template.New("test").Funcs(defaultTemplateFuncs).Parse(
		`{{date "2006-01-02 03:04:05 PM"}}`)
	assert.Nil(t, err)

	buff := bytes.NewBuffer(nil)
	assert.Nil(t, tmpl.Execute(buff, nil))
	fmt.Println(buff.String())
}

func TestDateDeltaFunc(t *testing.T) {
	tmpl, err := template.New("test").Funcs(defaultTemplateFuncs).Parse(
		`{{date_delta "2019-09-19" "2006-01-02" -1 1 1 }}`)
	assert.Nil(t, err)
	buff := bytes.NewBuffer(nil)
	assert.Nil(t, tmpl.Execute(buff, nil))
	ret := buff.String()
	assert.Equal(t, "2018-10-20", ret)
}

func TestTimestampFunc(t *testing.T) {
	tmpl, err := template.New("test").Funcs(defaultTemplateFuncs).Parse(
		`{{timestamp .Variable.precision}}`)
	assert.Nil(t, err)

	ctx := RenderContext{Variable: map[string]interface{}{"precision": "ms"}}
	buff := bytes.NewBuffer(nil)
	assert.Nil(t, tmpl.Execute(buff, ctx))
	assert.Equal(t, len(buff.String()), 13)
}

func TestPlusFunc(t *testing.T) {
	tmpl, err := template.New("test").Funcs(defaultTemplateFuncs).Parse(
		`{{ plus .Variable.string 2}}`)
	assert.Nil(t, err)
	ctx := RenderContext{Variable: map[string]interface{}{"string": "123", "int": 456, "float": 789.2}}

	buf := bytes.NewBuffer(nil)
	assert.Nil(t, tmpl.Execute(buf, ctx))
	assert.Equal(t, buf.String(), "125")

	tmpl, err = template.New("test").Funcs(defaultTemplateFuncs).Parse(
		`{{ plus .Variable.int 2}}`)
	assert.Nil(t, err)
	buf.Reset()
	assert.Nil(t, tmpl.Execute(buf, ctx))
	assert.Equal(t, buf.String(), "458")

	tmpl, err = template.New("test").Funcs(defaultTemplateFuncs).Parse(
		`{{ plus .Variable.float 2}}`)
	assert.Nil(t, err)
	buf.Reset()
	assert.Nil(t, tmpl.Execute(buf, ctx))
	assert.Equal(t, buf.String(), "791.2")
}

func TestGenRandString(t *testing.T) {
	tmpl, err := template.New("test").Funcs(defaultTemplateFuncs).Parse(
		`{{rand_string 8}}`)
	assert.Nil(t, err)

	buff := bytes.NewBuffer(nil)
	ctx := RenderContext{}
	assert.Nil(t, tmpl.Execute(buff, ctx))
	assert.Equal(t, len(buff.String()), 8)
}

func TestVarNameWithDash(t *testing.T) {
	p := struct {
		Data map[string]interface{}
	}{
		Data: map[string]interface{}{"deepmock-version": "v1.0.0"},
	}
	tp, err := template.New("test").Parse("{{.Data.deepmock-version}}")
	assert.NotNil(t, err)

	tp, err = template.New("test").Parse(
		`{{ $version := index .Data "deepmock-version"}}{{$version}}`)
	assert.Nil(t, err)

	buf := bytes.NewBuffer(nil)
	assert.Nil(t, tp.Execute(buf, p))
	assert.EqualValues(t, "v1.0.0", buf.String())
}

func TestRequestMatch_Match(t *testing.T) {
	var err error
	executor := &Executor{Method: []byte("GET")}
	executor.Path, err = regexp.Compile("/")
	assert.NoError(t, err)
	assert.True(t, executor.Match([]byte("/"), []byte("GET")))

	executor.Method = []byte("GET")
	executor.Path, err = regexp.Compile("/api/v1/create")
	assert.NoError(t, err)
	assert.True(t, executor.Match([]byte("/api/v1/create"), []byte("GET")))
	assert.False(t, executor.Match([]byte("/api/v1/create"), []byte("POST")))
	assert.False(t, executor.Match([]byte("/api/v1/update"), []byte("GET")))

	executor.Method = []byte("GET")
	executor.Path, err = regexp.Compile("/api/v[0-9]+/create")
	assert.NoError(t, err)
	assert.True(t, executor.Match([]byte("/api/v10/create"), []byte("GET")))
	assert.False(t, executor.Match([]byte("/api/va/create"), []byte("GET")))
}

func TestRuleExecutor_Minimal(t *testing.T) {
	rule := &Rule{
		Path:   "/api/v1/store/create",
		Method: "GET",
		Regulations: []*Regulation{
			{
				IsDefault: true,
				Template:  &Template{Body: `{"version": 1}`},
			}},
	}

	_, err := rule.To()
	assert.NoError(t, err)
}

func TestHandleHeaderTemplate(t *testing.T) {
	// 测试遍历header并修改的函数handleHeaderTemplate
	ctx := &fasthttp.RequestCtx{
		Request: fasthttp.Request{
			Header: fasthttp.RequestHeader{},
		},
	}
	ctx.Request.SetBody([]byte("{\"hello\":\"{{.Variable.name}}\", \"country\":\"{{.Query.country}}\",\"body\":\"{{.Form.nickname}}\"}"))
	ctx.Request.Header.SetRequestURIBytes([]byte("http://localhost:16600/redirect/baidu?redirect_uri=https%3A%2F%2Fwww.baidu.com&appid=appid&state=true"))

	// 模拟response header
	te := &TemplateExecutor{}
	te.header = &fasthttp.ResponseHeader{}
	te.header.Set("fake", "1")
	te.header.Set("rand-string", "{{rand_string 20}}")
	te.header.Set("user-agent", "Wechat")
	te.header.Set("uuid", "{{uuid}}")
	te.header.Set("Not-Exist-Func", "{{not_exist_func}}")
	te.header.Set("location", "{{.Query.redirect_uri}}?state={{.Query.state}}&app_id={{.Variable.app_id}}&auth_code={{.Variable.code}}")

	v := map[string]interface{}{
		"app_id": "app_id",
		"code":   "123456",
	}

	err := te.handleHeaderTemplate(ctx, v, nil)
	assert.Nil(t, err)

	fmt.Println("\n>>>>>>After render, the te.header is:")
	fmt.Println(string(te.header.Header()))
	decodeUrl, _ := url.QueryUnescape(string(te.header.Peek("Location")))
	assert.Equal(t, decodeUrl, "https://www.baidu.com?state=true&app_id=app_id&auth_code=123456")
	assert.Equal(t, string(te.header.Peek("not-exist-func")), "{{not_exist_func}}")
}

func TestParseParams(t *testing.T) {
	ctx := &fasthttp.RequestCtx{
		Request: fasthttp.Request{
			Header: fasthttp.RequestHeader{},
		},
	}

	ctx.Request.SetBody([]byte("{\"hello\":\"{{.Variable.name}}\", \"country\":\"{{.Query.country}}\",\"body\":\"{{.Form.nickname}}\"}"))
	ctx.Request.Header.SetRequestURIBytes([]byte("http://localhost:16600/redirect/baidu?redirect_uri=https%3A%2F%2Fwww.baidu.com&appid=appid&state=true"))

	v := map[string]interface{}{
		"app_id": "app_id",
		"code":   "123456",
	}
	rc := &RenderContext{}
	rc.parseParams(ctx, v, nil)

	assert.Equal(t, rc.Query["redirect_uri"], "https://www.baidu.com")
	assert.Equal(t, rc.Query["appid"], "appid")
	assert.Equal(t, rc.Query["state"], "true")
	assert.Equal(t, rc.Variable["app_id"], "app_id")
	assert.Equal(t, rc.Variable["code"], "123456")
}
