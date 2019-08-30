package deepmock

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qastub/deepmock/types"
)

func TestWeightingDice_Path(t *testing.T) {
	res := types.ResourceWeightingFactor{
		"SUCCESS": 1,
		"FAILED":  2,
	}
	wd := newWeighingDice(res)
	patch := types.ResourceWeightingFactor{"SUCCESS": 0}
	wd.patch(patch)
	assert.EqualValues(t, wd.total, 2)
	assert.Equal(t, wd.dice(), "FAILED")

	patch = types.ResourceWeightingFactor{"FAILED": 0, "UNKNOWN": 10}
	wd.patch(patch)
	assert.EqualValues(t, wd.total, 10)
	assert.Equal(t, wd.dice(), "UNKNOWN")
}
