package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wosai/deepmock/types/resource"
)

func TestWeightingDice_Path(t *testing.T) {
	res := resource.WeightingFactor{
		"SUCCESS": 1,
		"FAILED":  2,
	}
	wd := newWeighingDice(res)
	patch := resource.WeightingFactor{"SUCCESS": 0}
	wd.patch(patch)
	assert.EqualValues(t, wd.total, 2)
	assert.Equal(t, wd.dice(), "FAILED")

	patch = resource.WeightingFactor{"FAILED": 0, "UNKNOWN": 10}
	wd.patch(patch)
	assert.EqualValues(t, wd.total, 10)
	assert.Equal(t, wd.dice(), "UNKNOWN")
}

func TestWeightingPicker_Wrap(t *testing.T) {
	res := resource.Weight{
		"code":     resource.WeightingFactor{"CREATED": 1, "CLOSED": 2},
		"err_code": resource.WeightingFactor{"INVALID_NAME": 0, "INVALID_BANK_ACCOUNT": 2}}
	wfh := newWeightingPicker(res)
	wfh.wrap()

	assert.EqualValues(t, res, wfh.wrap())
}