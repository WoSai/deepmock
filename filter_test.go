package deepmock

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestHeaderFilter_Filter(t *testing.T) {
	hf := new(headerFilter)
	var params HeaderFilterParameters
	err := hf.withParameters(params)
	assert.Nil(t, err)

	assert.Equal(t, hf.mode, FilterModeAlwaysTrue)
	assert.True(t, hf.filter(nil))

	assert.Nil(t, hf.withParameters(HeaderFilterParameters{"mode": "exact", "content-type": "application/json"}))
	header := new(fasthttp.RequestHeader)
	header.SetContentType("application/json")
	assert.True(t, hf.filter(header))
	header.SetContentType("application/json; charset=utf-8")
	assert.False(t, hf.filter(header))

	assert.Nil(t, hf.withParameters(HeaderFilterParameters{"mode": "keyword", "content-type": "application/json"}))
	header = new(fasthttp.RequestHeader)
	header.SetContentType("application/json; charset=utf-8")
	assert.True(t, hf.filter(header))
	header.SetContentType("application/xml")
	assert.False(t, hf.filter(header))

	assert.Nil(t, hf.withParameters(HeaderFilterParameters{"mode": "regular", "authCode": "[0-9]+"}))
	header = new(fasthttp.RequestHeader)
	assert.False(t, hf.filter(header))
	header.Set("authCode", "123123")
	assert.True(t, hf.filter(header))
	header.Set("authCode", "hello world")
	assert.False(t, hf.filter(header))
}

func TestBodyFilter_Filter(t *testing.T) {
	bf := new(bodyFilter)
	var params BodyFilterParameters
	assert.Nil(t, bf.withParameters(params))
	assert.True(t, bf.filter(nil))

	assert.Nil(t, bf.withParameters(BodyFilterParameters{"mode": "keyword", "keyword": "foobar"}))
	assert.True(t, bf.filter([]byte(`hello foobar`)))
	assert.False(t, bf.filter([]byte(`hello world`)))

	assert.Nil(t, bf.withParameters(BodyFilterParameters{"mode": "regular", "regular": "[0-9]+"}))
	assert.False(t, bf.filter([]byte(`what's your mobile phone number'`)))
	assert.True(t, bf.filter([]byte(`my phone number is 110`)))
}

func TestQueryFilter_Filter(t *testing.T) {
	assertion := assert.New(t)

	qf := new(queryFilter)
	var params QueryFilterParameters
	assertion.Nil(qf.withParameters(params))
	assertion.True(qf.filter(nil))

	assertion.Nil(qf.withParameters(QueryFilterParameters{"mode": "exact", "nation": "中国"}))
	query := new(fasthttp.Args)
	query.Set("version", "1")
	query.Set("nation", "中国")
	assertion.True(qf.filter(query))
	query.Set("nation", "USA")
	assertion.False(qf.filter(query))
	query.Del("nation")
	assertion.False(qf.filter(query))

	assertion.Nil(qf.withParameters(QueryFilterParameters{"mode": "keyword", "nation": "中国"}))
	query.Set("nation", "CHINA")
	assertion.False(qf.filter(query))
	query.Set("nation", "中国123")
	assertion.True(qf.filter(query))

	assertion.Nil(qf.withParameters(QueryFilterParameters{"mode": "regular", "age": "[0-9]+"}))
	query.Set("age", "18")
	assertion.True(qf.filter(query))
	query.Set("age", "unknown")
	assertion.False(qf.filter(query))
}
