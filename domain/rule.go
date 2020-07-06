package domain

import (
	"errors"

	"github.com/wosai/deepmock/misc"
)

type (
	Rule struct {
		ID          string
		Path        string
		Method      string
		Variable    map[string]interface{}
		Weight      map[string]map[string]uint
		Regulations []*Regulation
		Version     int
	}

	Regulation struct {
		IsDefault bool
		Filter    *Filter
		Template  *Template
	}

	Filter struct {
		Query  map[string]string
		Header map[string]string
		Body   map[string]string
	}

	Template struct {
		IsTemplate     bool
		Header         map[string]string
		StatusCode     int
		Body           string
		B64EncodedBody string
	}
)

func (f *Filter) Validate() error {
	if f.Header != nil {
		if _, ok := f.Header["mode"]; !ok {
			return errors.New("missing mode in header filter")
		}
	}

	if f.Query != nil {
		if _, ok := f.Query["mode"]; !ok {
			return errors.New("missing mode in query filter")
		}
	}

	if f.Body != nil {
		if _, ok := f.Body["mode"]; !ok {
			return errors.New("missing mode in body filter")
		}
	}
	return nil
}

func (r *Regulation) Validate() error {
	if !r.IsDefault && r.Filter == nil {
		return errors.New("unreachable regulation")
	}
	if err := r.Filter.Validate(); err != nil {
		return err
	}
	if r.Template == nil {
		return errors.New("missing response template")
	}
	return nil
}

// Validate 校验Rule的有效性
func (rule *Rule) Validate() error {
	if rule.ID != "" && misc.GenID([]byte(rule.Path), []byte(rule.Method)) != rule.ID {
		return errors.New("invalid rule id")
	}
	if len(rule.Path) == 0 {
		return errors.New("bad rule path")
	}
	if len(rule.Method) == 0 {
		return errors.New("bad rule method")
	}
	if len(rule.Regulations) == 0 {
		return errors.New("missing regulation")
	}

	var d int
	for _, reg := range rule.Regulations {
		if reg.IsDefault {
			d++
		}
		if err := reg.Validate(); err != nil {
			return err
		}
	}
	if d != 1 {
		return errors.New("no default regulation or provided more than one")
	}
	return nil
}

// SupplyID 补充对象ID，如果不存在的话
func (rule *Rule) SupplyID() (string, bool) {
	if rule.ID != "" {
		return rule.ID, false
	}

	rule.ID = misc.GenID([]byte(rule.Path), []byte(rule.Method))
	return rule.ID, true
}

// Patch 更新对象
func (rule *Rule) Patch(nr *Rule) error {
	rule.Version++

	// variable
	switch {
	case rule.Variable == nil && nr.Variable != nil:
		rule.Variable = nr.Variable

	case rule.Variable != nil && nr.Variable != nil:
		for k, v := range nr.Variable {
			rule.Variable[k] = v
		}

	default:
	}

	// weight
	switch {
	case rule.Weight == nil && nr.Weight != nil:
		rule.Weight = nr.Weight

	case rule.Weight != nil && nr.Weight != nil:
		for k, factor := range nr.Weight {
			current, exists := rule.Weight[k]
			if exists {
				for ele, v := range factor {
					current[ele] = v
				}
			} else {
				rule.Weight[k] = factor
			}
		}
	default:
	}

	// regulation
	if len(nr.Regulations) > 0 {
		rule.Regulations = nr.Regulations
	}

	return rule.Validate()
}

func (rule *Rule) Put(nr *Rule) error {
	rule.Version++

	rule.Variable = nr.Variable
	rule.Weight = nr.Weight
	rule.Regulations = nr.Regulations
	return rule.Validate()
}
