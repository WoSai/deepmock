package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePathVar(t *testing.T) {
	path := []byte("/api/v1/rule")
	uri := []byte("/api/v1/rule/123")

	assert.Equal(t, parsePathVar(path, uri), "123")
}
