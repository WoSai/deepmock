package mock

import (
	"math/rand"

	"github.com/wosai/deepmock/types/resource"
)

type (
	weightingDice struct {
		total   uint
		storage []string
		raw     resource.WeightingFactor
	}

	weightingPicker map[string]*weightingDice
)

func newWeighingDice(res resource.WeightingFactor) *weightingDice {
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

func (wp *weightingDice) patch(res resource.WeightingFactor) {
	for nk, nv := range res {
		wp.raw[nk] = nv
	}

	wp.preDistribute()
}

func newWeightingPicker(res resource.Weight) weightingPicker {
	wfh := make(weightingPicker)
	for k, v := range res {
		wfh[k] = newWeighingDice(v)
	}
	return wfh
}

func (wfg weightingPicker) wrap() resource.Weight {
	if wfg == nil {
		return nil
	}

	ret := make(resource.Weight)
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

func (wfg weightingPicker) patch(res resource.Weight) {
	for k, v := range res {
		d, ok := wfg[k]
		if ok {
			d.patch(v)
		} else {
			wfg[k] = newWeighingDice(v)
		}
	}
}
