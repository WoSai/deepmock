package service

import (
	"errors"

	"github.com/wosai/deepmock/types/resource"
)

func ValidateRule(rule *resource.Rule) error {
	if rule.Path == "" {
		return errors.New("missing mock api path")
	}
	if rule.Method == "" {
		return errors.New("missing mock api method")
	}
	if rule.Responses == nil {
		return errors.New("missing response regulations")
	}

	return validateResponseRegulationSet(rule.Responses)
}

func validateResponseRegulation(rr *resource.ResponseRegulation) error {
	if !rr.IsDefault && rr.Filter == nil {
		return errors.New("missing filter rule, or set as default response")
	}
	if rr.Response == nil {
		return errors.New("missing response template")
	}
	return nil
}

func validateResponseRegulationSet(rrs resource.ResponseRegulationSet) error {
	var d int
	if rrs == nil {
		return errors.New("missing mock response")
	}

	for _, r := range rrs {
		if r.IsDefault {
			d++
		}
		if err := validateResponseRegulation(r); err != nil {
			return err
		}
	}
	if d != 1 {
		return errors.New("no default response or provided more than one")
	}
	return nil
}

func validateHeaderFilterParameters(hfp resource.HeaderFilterParameters) error {
	if hfp == nil {
		return nil
	}

	if _, ok := hfp["mode"]; !ok {
		return errors.New("missing filter mode")
	}
	return nil
}

func validateQueryFilterParameters(qfp resource.QueryFilterParameters) error {
	if qfp == nil {
		return nil
	}

	if _, ok := qfp["mode"]; !ok {
		return errors.New("missing filter mode")
	}
	return nil
}

func validateBodyFilterParameters(bfp resource.BodyFilterParameters) error {
	if bfp == nil {
		return nil
	}

	if _, ok := bfp["mode"]; !ok {
		return errors.New("missing filter mode")
	}
	return nil
}