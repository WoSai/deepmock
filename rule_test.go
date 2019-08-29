package deepmock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestMatch_Match(t *testing.T) {
	rm := &RequestMatch{Path: "/", Method: "Get"}
	assert.True(t, rm.Match("GET", "/"))

	rm = &RequestMatch{Method: "GET", Path: "/api/v1/create"}
	assert.False(t, rm.Match("POST", "/api/v1/create"))

	rm = &RequestMatch{Method: "GET", Path: "/api/v1/create"}
	assert.False(t, rm.Match("GET", "/api/v1/update"))

	rm = &RequestMatch{Method: "GET", Path: "/api/v[0-9]+/create"}
	assert.True(t, rm.Match("GET", "/api/v10/create"))
	assert.False(t, rm.Match("GET", "/api/va/create"))
}
