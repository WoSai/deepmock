package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
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
