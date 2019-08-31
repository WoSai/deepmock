package deepmock

import (
	"testing"

	"github.com/qastub/deepmock/types"
	"github.com/stretchr/testify/assert"
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

func TestWeightingPicker_Wrap(t *testing.T) {
	res := types.ResourceWeight{
		"code":     types.ResourceWeightingFactor{"CREATED": 1, "CLOSED": 2},
		"err_code": types.ResourceWeightingFactor{"INVALID_NAME": 0, "INVALID_BANK_ACCOUNT": 2}}
	wfh := newWeightingPicker(res)
	wfh.wrap()

	assert.EqualValues(t, res, wfh.wrap())
}
