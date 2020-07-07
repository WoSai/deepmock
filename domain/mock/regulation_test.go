package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wosai/deepmock/types/resource"
)

func TestResponseRegulation_Wrap(t *testing.T) {
	res := &resource.ResponseRegulation{
		IsDefault: true,
		Filter: &resource.Filter{
			Body: resource.BodyFilterParameters{"mode": "keyword", "keyword": "createStore"},
		},
		Response: &resource.ResponseTemplate{Body: "hello pingpong", Header: resource.HeaderTemplate{"Content-Type": "text/plaintext"}},
	}
	d1, _ := domain.json.Marshal(res)

	rr, err := newResponseRegulation(res)
	assert.Nil(t, err)
	d2, _ := domain.json.Marshal(rr.wrap())
	assert.Equal(t, string(d1), string(d2))
}
