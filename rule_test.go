package deepmock

import (
	"testing"

	"github.com/qastub/deepmock/types"

	"github.com/stretchr/testify/assert"
)

func TestRequestMatch_Match(t *testing.T) {
	rm, err := newRequestMatcher(&types.ResourceRequestMatcher{Path: "/", Method: "Get"})
	assert.Nil(t, err)
	assert.True(t, rm.match([]byte("/"), []byte("GET")))

	rm, err = newRequestMatcher(&types.ResourceRequestMatcher{Method: "GET", Path: "/api/v1/create"})
	assert.Nil(t, err)
	assert.False(t, rm.match([]byte("/api/v1/create"), []byte("POST")))

	rm, err = newRequestMatcher(&types.ResourceRequestMatcher{Method: "GET", Path: "/api/v1/create"})
	assert.Nil(t, err)
	assert.False(t, rm.match([]byte("/api/v1/update"), []byte("GET")))

	rm, err = newRequestMatcher(&types.ResourceRequestMatcher{Method: "GET", Path: "/api/v[0-9]+/create"})
	assert.Nil(t, err)
	assert.True(t, rm.match([]byte("/api/v10/create"), []byte("GET")))
	assert.False(t, rm.match([]byte("/api/va/create"), []byte("GET")))
}

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

func TestResponseRegulation_Wrap(t *testing.T) {
	res := &types.ResourceResponseRegulation{
		IsDefault: true,
		Filter: &types.ResourceFilter{
			Body: types.ResourceBodyFilterParameters{"mode": "keyword", "keyword": "createStore"},
		},
		Response: &types.ResourceResponseTemplate{Body: "hello pingpong", Header: types.ResourceHeaderTemplate{"Content-Type": "text/plaintext"}},
	}
	d1, _ := json.Marshal(res)

	rr, err := newResponseRegulation(res)
	assert.Nil(t, err)
	d2, _ := json.Marshal(rr.wrap())
	assert.Equal(t, string(d1), string(d2))
}

func TestWeightingFactorHub_Wrap(t *testing.T) {
	res := types.ResourceWeight{
		"code":     types.ResourceWeightingFactor{"CREATED": 1, "CLOSED": 2},
		"err_code": types.ResourceWeightingFactor{"INVALID_NAME": 0, "INVALID_BANK_ACCOUNT": 2}}
	wfh := newWeightingPicker(res)
	wfh.wrap()

	assert.EqualValues(t, res, wfh.wrap())
}
