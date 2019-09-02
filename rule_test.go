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

func TestRuleManager_FindExecutor(t *testing.T) {
	rm := newRuleManager()

	r1 := &types.ResourceRule{
		Request: &types.ResourceRequestMatcher{Path: "/api/v1/rule/[0-9]+", Method: "GET"},
		Responses: types.ResourceResponseRegulationSet{&types.ResourceResponseRegulation{
			IsDefault: true,
			Response:  &types.ResourceResponseTemplate{Body: "hello rule"},
		}},
	}

	r2 := &types.ResourceRule{
		Request: &types.ResourceRequestMatcher{Path: "/api/v1/store/[0-9]+", Method: "GET"},
		Responses: types.ResourceResponseRegulationSet{&types.ResourceResponseRegulation{
			IsDefault: true,
			Response:  &types.ResourceResponseTemplate{Body: "hello store"},
		}},
	}

	re1, err := rm.createRule(r1)
	assert.Nil(t, err)
	assert.NotNil(t, re1)

	re2, err := rm.createRule(r2)
	assert.Nil(t, err)
	assert.NotNil(t, re2)

	_, _, cached := rm.findExecutor([]byte(`/api/v1/rule/123`), []byte(`GET`))
	assert.False(t, cached)

	_, _, cached = rm.findExecutor([]byte(`/api/v1/rule/123`), []byte(`GET`))
	assert.True(t, cached)

	_, _, cached = rm.findExecutor([]byte(`/api/v1/store/123`), []byte(`GET`))
	assert.False(t, cached)

	_, _, cached = rm.findExecutor([]byte(`/api/v1/store/123`), []byte(`GET`))
	assert.True(t, cached)

	r1.ID = re1.id()
	rm.deleteRule(r1)

	_, founded, cached := rm.findExecutor([]byte(`/api/v1/rule/123`), []byte(`GET`))
	assert.False(t, cached)
	assert.False(t, founded)
}

func TestRuleManager_UpdateRule(t *testing.T) {
	rm := newRuleManager()

	r := &types.ResourceRule{
		Request: &types.ResourceRequestMatcher{Path: "/api/v1/rule/[0-9]+", Method: "GET"},
		Responses: types.ResourceResponseRegulationSet{&types.ResourceResponseRegulation{
			IsDefault: true,
			Response:  &types.ResourceResponseTemplate{Body: "hello rule"},
		}},
	}

	//r2 := &types.ResourceRule{
	//	Request: &types.ResourceRequestMatcher{Path: "/api/v1/store/[0-9]+", Method: "GET"},
	//	Responses: types.ResourceResponseRegulationSet{&types.ResourceResponseRegulation{
	//		IsDefault: true,
	//		Response:  &types.ResourceResponseTemplate{Body: "hello store"},
	//	}},
	//}

	re1, err := rm.createRule(r)
	assert.Nil(t, err)
	assert.NotNil(t, re1)

	r.Request.Method = "POST"
	_, err = rm.updateRule(r)
	assert.Error(t, err, "the rule to update is not exists")
	re3, _ := rm.getRuleByID(re1.id())
	assert.Equal(t, re1, re3)

	r.Request.Method = "GET"
	r.Responses = append(r.Responses, &types.ResourceResponseRegulation{
		Filter:   &types.ResourceFilter{Body: types.ResourceBodyFilterParameters{"mode": "always_true"}},
		Response: &types.ResourceResponseTemplate{Body: "foobar"}})
	re4, err := rm.updateRule(r)
	assert.Nil(t, err)
	assert.NotEqual(t, re1, re4)
}
