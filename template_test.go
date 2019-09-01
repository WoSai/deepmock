package deepmock

import (
	"bytes"
	"fmt"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
