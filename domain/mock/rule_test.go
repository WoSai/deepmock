package mock

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wosai/deepmock/types"
)

func TestRequestMatch_Match(t *testing.T) {
	rm, err := newRequestMatcher("/", "GET")
	assert.Nil(t, err)
	assert.True(t, rm.match([]byte("/"), []byte("GET")))

	rm, err = newRequestMatcher("/api/v1/create", "GET")
	assert.Nil(t, err)
	assert.False(t, rm.match([]byte("/api/v1/create"), []byte("POST")))

	rm, err = newRequestMatcher("/api/v1/create", "GET")
	assert.Nil(t, err)
	assert.False(t, rm.match([]byte("/api/v1/update"), []byte("GET")))

	rm, err = newRequestMatcher("/api/v[0-9]+/create", "GET")
	assert.Nil(t, err)
	assert.True(t, rm.match([]byte("/api/v10/create"), []byte("GET")))
	assert.False(t, rm.match([]byte("/api/va/create"), []byte("GET")))
}

func TestRuleExecutor_New(t *testing.T) {
	rule := &types.ResourceRule{
		Path:   "/api/v1/store/create",
		Method: "GET",
		Responses: &types.ResourceResponseRegulationSet{
			&types.ResourceResponseRegulation{IsDefault: true, Response: &types.ResourceResponseTemplate{Body: `{"version": 1}`}},
		},
	}

	_, err := newRuleExecutor(rule)
	assert.Nil(t, err)
}

func TestRuleExecutor_Wrap(t *testing.T) {
	rule := &types.ResourceRule{
		Method:   "GET",
		Path:     "/api/v1/store/create",
		Variable: &types.ResourceVariable{"version": 1, "name": "foobar"},
		Weight:   &types.ResourceWeight{"return_code": types.ResourceWeightingFactor{"success": 1, "failed": 2}, "error_code": types.ResourceWeightingFactor{"invalid_name": 100}},
		Responses: &types.ResourceResponseRegulationSet{
			&types.ResourceResponseRegulation{IsDefault: true, Filter: nil, Response: &types.ResourceResponseTemplate{Body: `{"version": 1}`}},
		},
	}

	data, _ := deepmock.json.Marshal(rule)

	re, err := newRuleExecutor(rule)
	assert.Nil(t, err)
	assert.NotEmpty(t, re.id())

	rule2 := re.wrap()
	rule2.ID = ""
	data2, _ := deepmock.json.Marshal(rule2)
	assert.Equal(t, data, data2)
}

func TestRuleExecutor_Path(t *testing.T) {
	rule := &types.ResourceRule{
		Path:     "/api/v1/rule/[0-9]+",
		Method:   "GET",
		Variable: &types.ResourceVariable{"version": 1},
		Weight:   &types.ResourceWeight{"code": types.ResourceWeightingFactor{"S": 1, "F": 2}},
		Responses: &types.ResourceResponseRegulationSet{&types.ResourceResponseRegulation{
			IsDefault: true,
			Response:  &types.ResourceResponseTemplate{Body: "hello rule"},
		}},
	}

	re, err := newRuleExecutor(rule)
	assert.Nil(t, err)

	p1 := &types.ResourceRule{
		ID:       re.id(),
		Variable: &types.ResourceVariable{"version": 2},
		Weight: &types.ResourceWeight{
			"code":   types.ResourceWeightingFactor{"F": 3},
			"result": types.ResourceWeightingFactor{"SUCCESS": 0, "FAILED": 10}},
	}
	assert.Nil(t, re.patch(p1))

	rule = re.wrap()
	assert.EqualValues(t,
		*rule.Weight,
		types.ResourceWeight{
			"code":   types.ResourceWeightingFactor{"S": 1, "F": 3},
			"result": types.ResourceWeightingFactor{"SUCCESS": 0, "FAILED": 10},
		})

	assert.EqualValues(t,
		*rule.Variable,
		types.ResourceVariable{"version": 2})

	p2 := &types.ResourceRule{
		ID: re.id(),
		Responses: &types.ResourceResponseRegulationSet{&types.ResourceResponseRegulation{
			IsDefault: true,
			Response:  &types.ResourceResponseTemplate{Body: "foobar"},
		}},
	}

	assert.Nil(t, re.patch(p2))
	rs := *re.wrap().Responses
	assert.Equal(t, rs[0].Response.Body, "foobar")
}

func TestRuleManager_FindExecutor(t *testing.T) {
	rm := newRuleManager()

	r1 := &types.ResourceRule{
		Path:   "/api/v1/rule/[0-9]+",
		Method: "GET",
		Responses: &types.ResourceResponseRegulationSet{&types.ResourceResponseRegulation{
			IsDefault: true,
			Response:  &types.ResourceResponseTemplate{Body: "hello rule"},
		}},
	}

	r2 := &types.ResourceRule{
		Path:   "/api/v1/store/[0-9]+",
		Method: "GET",
		Responses: &types.ResourceResponseRegulationSet{&types.ResourceResponseRegulation{
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

func TestRuleManager_PathRule(t *testing.T) {
	rm := newRuleManager()
	r := &types.ResourceRule{
		Path:   "/api/v1/rule/[0-9]+",
		Method: "GET",
		Responses: &types.ResourceResponseRegulationSet{&types.ResourceResponseRegulation{
			IsDefault: true,
			Response:  &types.ResourceResponseTemplate{Body: "hello rule"},
		}},
	}
	re, err := rm.createRule(r)
	assert.Nil(t, err)

	_, err = rm.patchRule(&types.ResourceRule{ID: "123"})
	assert.NotNil(t, err)

	_, err = rm.patchRule(&types.ResourceRule{ID: re.id()})
	assert.Nil(t, err)
}

func TestRuleManager_UpdateRule(t *testing.T) {
	rm := newRuleManager()

	r := &types.ResourceRule{
		Path:   "/api/v1/rule/[0-9]+",
		Method: "GET",
		Responses: &types.ResourceResponseRegulationSet{&types.ResourceResponseRegulation{
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

	r.Method = "POST"
	_, err = rm.updateRule(r)
	assert.Error(t, err, "the rule to update is not exists")
	re3, _ := rm.getRuleByID(re1.id())
	assert.Equal(t, re1, re3)

	r.Method = "GET"
	*r.Responses = append(*r.Responses, &types.ResourceResponseRegulation{
		Filter:   &types.ResourceFilter{Body: types.ResourceBodyFilterParameters{"mode": "always_true"}},
		Response: &types.ResourceResponseTemplate{Body: "foobar"}})
	re4, err := rm.updateRule(r)
	assert.Nil(t, err)
	assert.NotEqual(t, re1, re4)
}

func BenchmarkRuleManager_DataRace(b *testing.B) {
	rm := newRuleManager()

	r1 := &types.ResourceRule{
		Path:   "/api/v1/rule/[0-9]+",
		Method: "GET",
		Responses: &types.ResourceResponseRegulationSet{&types.ResourceResponseRegulation{
			IsDefault: true,
			Response:  &types.ResourceResponseTemplate{Body: "hello rule"},
		}},
	}

	r2 := &types.ResourceRule{
		Path:   "/api/v1/store/[0-9]+",
		Method: "GET",
		Responses: &types.ResourceResponseRegulationSet{&types.ResourceResponseRegulation{
			IsDefault: true,
			Response:  &types.ResourceResponseTemplate{Body: "hello store"},
		}},
	}

	re1, err := rm.createRule(r1)
	assert.Nil(b, err)
	_, err = rm.createRule(r2)
	assert.Nil(b, err)

	patch := &types.ResourceRule{
		ID:       re1.id(),
		Variable: &types.ResourceVariable{"version": 1},
		Weight:   &types.ResourceWeight{"code": types.ResourceWeightingFactor{"S": 1, "F": 2}},
		Responses: &types.ResourceResponseRegulationSet{&types.ResourceResponseRegulation{
			IsDefault: true,
			Response:  &types.ResourceResponseTemplate{Body: "hello patch"},
		}},
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			switch rand.Intn(10) {
			case 0, 1, 2, 3, 4:
				rm.findExecutor([]byte("/api/v1/rule/123"), []byte(`GET`))
			case 5, 6, 7, 8:
				rm.findExecutor([]byte("/api/v1/store/123"), []byte(`GET`))
			default:
				_, err := rm.patchRule(patch)
				assert.Nil(b, err)
			}
		}
	})

}
