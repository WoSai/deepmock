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

func TestRuleExecutor_New(t *testing.T) {
	rule := &types.ResourceRule{
		Request: &types.ResourceRequestMatcher{Method: "GET", Path: "/api/v1/store/create"},
		Responses: types.ResourceResponseRegulationSet{
			&types.ResourceResponseRegulation{IsDefault: true, Response: &types.ResourceResponseTemplate{Body: `{"version": 1}`}},
		},
	}

	_, err := newRuleExecutor(rule)
	assert.Nil(t, err)
}

func TestRuleExecutor_Wrap(t *testing.T) {
	rule := &types.ResourceRule{
		Request:  &types.ResourceRequestMatcher{Method: "GET", Path: "/api/v1/store/create"},
		Variable: types.ResourceVariable{"version": 1, "name": "foobar"},
		Weight:   types.ResourceWeight{"return_code": types.ResourceWeightingFactor{"success": 1, "failed": 2}, "error_code": types.ResourceWeightingFactor{"invalid_name": 100}},
		Responses: types.ResourceResponseRegulationSet{
			&types.ResourceResponseRegulation{IsDefault: true, Filter: nil, Response: &types.ResourceResponseTemplate{Body: `{"version": 1}`}},
		},
	}

	data, _ := json.Marshal(rule)

	re, err := newRuleExecutor(rule)
	assert.Nil(t, err)
	assert.NotEmpty(t, re.id())

	rule2 := re.wrap()
	rule2.ID = ""
	data2, _ := json.Marshal(rule2)
	assert.Equal(t, data, data2)
}
