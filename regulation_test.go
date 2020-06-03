package deepmock

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wosai/deepmock/types"
)

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
