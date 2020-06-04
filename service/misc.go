package service

import (
	"strings"

	"github.com/wosai/deepmock"
	"github.com/wosai/deepmock/types/entity"
	"github.com/wosai/deepmock/types/resource"
)

func convertAsEntity(rule *resource.Rule) (*entity.Rule, error) {
	re := &entity.Rule{
		Path:     rule.Path,
		Method:   strings.ToUpper(rule.Method),
		Disabled: rule.Disabled,
	}
	if rule.ID == "" {
		rule.ID = deepmock.GenID([]byte(re.Path), []byte(re.Method))
	}
	re.ID = rule.ID
	if rule.Variable != nil {

	}
	if rule.Variable != nil {
		if data, err := json.Marshal(rule.Variable); err != nil {
			return nil, err
		} else {
			re.Variable = data
		}
	}
	if rule.Weight != nil {
		if data, err := json.Marshal(rule.Weight); err != nil {
			return nil, err
		} else {
			re.Weight = data
		}
	}
	if rule.Responses != nil {
		if data, err := json.Marshal(rule.Weight); err != nil {
			return nil, err
		} else {
			re.Responses = data
		}
	}
	if !rule.CreatedTime.IsZero() {
		re.CreatedTime = rule.CreatedTime
	}
	if !rule.ModifiedTime.IsZero() {
		re.ModifiedTime = rule.ModifiedTime
	}
	if rule.Version != 0 {
		re.Version = rule.Version
	}
	return re, nil
}

func convertAsResource(rule *entity.Rule) (*resource.Rule, error) {
	rl := &resource.Rule{
		ID:           rule.ID,
		Path:         rule.Path,
		Method:       rule.Method,
		Version:      rule.Version,
		CreatedTime:  rule.CreatedTime,
		ModifiedTime: rule.ModifiedTime,
		Disabled:     rule.Disabled,
	}
	var err error
	if rule.Variable != nil {
		v := new(resource.Variable)
		if err = json.Unmarshal(rule.Variable, v); err != nil {
			return nil, err
		}
	}

	if rule.Weight != nil {
		w := new(resource.Weight)
		if err = json.Unmarshal(rule.Weight, w); err != nil {
			return nil, err
		}
	}

	if rule.Responses != nil {
		r := new(resource.ResponseRegulationSet)
		if err = json.Unmarshal(rule.Responses, r); err != nil {
			return nil, err
		}
	}
	return rl, nil
}
