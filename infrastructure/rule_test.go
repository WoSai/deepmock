package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wosai/deepmock/types"
)

func TestConvertRuleDO(t *testing.T) {
	do := &types.RuleDO{
		ID:        "098e0b26",
		Path:      "/rpc/pushTarget",
		Method:    "POST",
		Variable:  nil,
		Weight:    nil,
		Responses: []byte(`[{"is_default":true,"response":{"header":{"Content-Type":"application/json"},"body":"{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":null}"}}]`),
	}

	entity, err := convertRuleDO(do)
	assert.NoError(t, err)
	assert.Equal(t, len(entity.Regulations), 1)
	assert.True(t, entity.Regulations[0].IsDefault)
	assert.NoError(t, entity.Validate())
}
