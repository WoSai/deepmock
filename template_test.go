package deepmock

import (
	"bytes"
	"fmt"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wosai/deepmock/types"
)

func TestNewResponseTemplate(t *testing.T) {
	res := &types.ResourceResponseTemplate{
		IsTemplate:     true,
		Header:         types.ResourceHeaderTemplate{"Content-Type": "application/json", "Authorization": "123123"},
		StatusCode:     500,
		Body:           "hello world",
		B64EncodedBody: "aGVsbG8gZm9vYmFyIQ==",
	}
	rt, err := newResponseTemplate(res)
	assert.Nil(t, err)
	assert.EqualValues(t, res, rt.raw)
	assert.True(t, rt.isTemplate)
	assert.True(t, rt.isBinData)
	assert.Equal(t, rt.body, []byte("hello foobar!"))
	assert.NotNil(t, rt.htmlTemplate)
	assert.Equal(t, rt.header.StatusCode(), 500)
	assert.Equal(t, rt.header.ContentType(), []byte("application/json"))
	assert.Equal(t, rt.header.Peek("Authorization"), []byte("123123"))
}

func TestUUIDFunc(t *testing.T) {
	text := `{{uuid}}`
	tmpl, err := template.New("test").Funcs(defaultTemplateFuncs).Parse(text)
	assert.Nil(t, err)

	buff := bytes.NewBuffer(nil)
	ctx := renderContext{}
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

	ctx := renderContext{Variable: ruleVariable{"precision": "ms"}}
	buff := bytes.NewBuffer(nil)
	assert.Nil(t, tmpl.Execute(buff, ctx))
	assert.Equal(t, len(buff.String()), 13)
}

func TestPlusFunc(t *testing.T) {
	tmpl, err := template.New("test").Funcs(defaultTemplateFuncs).Parse(
		`{{ plus .Variable.string 2}}`)
	assert.Nil(t, err)
	ctx := renderContext{Variable: ruleVariable{"string": "123", "int": 456, "float": 789.2}}

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

func TestGenRandstring(t *testing.T) {
	tmpl, err := template.New("test").Funcs(defaultTemplateFuncs).Parse(
		`{{rand_string 8}}`)
	assert.Nil(t, err)

	buff := bytes.NewBuffer(nil)
	ctx := renderContext{}
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
