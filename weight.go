package deepmock

import (
	"math/rand"

	"github.com/wosai/deepmock/types"
)

type (
	weightingDice struct {
		total   uint
		storage []string
		raw     types.ResourceWeightingFactor
	}

	weightingPicker map[string]*weightingDice
)

func newWeighingDice(res types.ResourceWeightingFactor) *weightingDice {
	wp := &weightingDice{raw: res}
	wp.preDistribute()
	return wp
}

func (wp *weightingDice) preDistribute() {
	wp.total = 0
	wp.storage = []string{}
	for k, v := range wp.raw {
		for i := 0; i < int(v); i++ {
			wp.storage = append(wp.storage, k)
			wp.total++
		}
	}
}

func (wp *weightingDice) dice() string {
	return wp.storage[rand.Intn(int(wp.total))]
}

func (wp *weightingDice) patch(res types.ResourceWeightingFactor) {
	for nk, nv := range res {
		wp.raw[nk] = nv
	}

	wp.preDistribute()
}

func newWeightingPicker(res types.ResourceWeight) weightingPicker {
	wfh := make(weightingPicker)
	for k, v := range res {
		wfh[k] = newWeighingDice(v)
	}
	return wfh
}

func (wfg weightingPicker) wrap() types.ResourceWeight {
	if wfg == nil {
		return nil
	}

	ret := make(types.ResourceWeight)
	for k, v := range wfg {
		ret[k] = v.raw
	}
	return ret
}

func (wfg weightingPicker) dice() params {
	p := make(params)
	for k, w := range wfg {
		p[k] = w.dice()
	}
	return p
}

func (wfg weightingPicker) patch(res types.ResourceWeight) {
	for k, v := range res {
		d, ok := wfg[k]
		if ok {
			d.patch(v)
		} else {
			wfg[k] = newWeighingDice(v)
		}
	}
}
