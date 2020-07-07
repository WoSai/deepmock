package domain

import (
	"encoding/base64"
	"errors"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/wosai/deepmock/misc"
)

type (
	Rule struct {
		ID          string
		Path        string
		Method      string
		Variable    map[string]interface{}
		Weight      map[string]WeightFactor
		Regulations []*Regulation
		Version     int
	}

	Regulation struct {
		IsDefault bool
		Filter    *Filter
		Template  *Template
	}

	Filter struct {
		Query  QueryFilterParams
		Header HeaderFilterParams
		Body   BodyFilterParams
	}

	Template struct {
		IsTemplate     bool
		Header         map[string]string
		StatusCode     int
		Body           string
		B64EncodedBody string
	}

	WeightFactor       map[string]uint
	QueryFilterParams  map[string]string
	HeaderFilterParams map[string]string
	BodyFilterParams   map[string]string
)

func (f *Filter) Validate() error {
	if f == nil {
		return nil
	}
	if f.Header != nil {
		if _, ok := f.Header[ModeField]; !ok {
			return errors.New("missing mode in header filter")
		}
	}

	if f.Query != nil {
		if _, ok := f.Query[ModeField]; !ok {
			return errors.New("missing mode in query filter")
		}
	}

	if f.Body != nil {
		if _, ok := f.Body[ModeField]; !ok {
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
	if r.Template.StatusCode == 0 {
		r.Template.StatusCode = http.StatusOK
	}
	return nil
}

func (r *Regulation) To() (*RegulationExecutor, error) {
	var err error

	exec := &RegulationExecutor{
		IsDefault: r.IsDefault,
		Filter:    new(FilterExecutor),
		Template:  new(TemplateExecutor),
	}
	if r.Filter != nil {
		exec.Filter.Query, err = r.Filter.Query.To()
		if err != nil {
			return nil, err
		}

		exec.Filter.Header, err = r.Filter.Header.To()
		if err != nil {
			return nil, err
		}

		exec.Filter.Body, err = r.Filter.Body.To()
		if err != nil {
			return nil, err
		}
	}

	exec.Template, err = r.Template.To()
	if err != nil {
		return nil, err
	}
	return exec, nil
}

// Validate 校验Rule的有效性
func (rule *Rule) Validate() error {
	rule.Method = strings.ToUpper(rule.Method)
	rule.SupplyID()

	if rule.ID != "" && misc.GenID([]byte(rule.Path), []byte(rule.Method)) != rule.ID {
		return errors.New("invalid rule id")
	}
	if len(rule.Path) == 0 {
		return errors.New("bad rule Path")
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

	// Variable
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

func (rule *Rule) To() (*Executor, error) {
	if err := rule.Validate(); err != nil {
		return nil, err
	}
	var err error
	exec := &Executor{
		ID:          rule.ID,
		Method:      []byte(rule.Method),
		Variable:    rule.Variable,
		Regulations: nil,
		Version:     rule.Version,
	}
	exec.Path, err = regexp.Compile(rule.Path)
	if err != nil {
		return nil, err
	}
	exec.Weight = make(WeightPicker, len(rule.Weight))
	for k, factor := range rule.Weight {
		exec.Weight[k] = factor.To()
	}

	exec.Regulations = make([]*RegulationExecutor, len(rule.Regulations))
	for index, regulation := range rule.Regulations {
		re, err := regulation.To()
		if err != nil {
			return nil, err
		}
		exec.Regulations[index] = re
	}
	return nil, nil
}

func (wf WeightFactor) To() *WeightDice {
	wd := &WeightDice{
		total:        0,
		distribution: []string{},
		factor:       wf,
	}

	for k, v := range wf {
		for i := 0; i < int(v); i++ {
			wd.distribution = append(wd.distribution, k)
			wd.total++
		}
	}
	return wd
}

func (qfp QueryFilterParams) To() (*QueryFilterExecutor, error) {
	if qfp == nil {
		return &QueryFilterExecutor{mode: FilterModeAlwaysTrue}, nil
	}

	mode := qfp[ModeField]
	qfe := &QueryFilterExecutor{
		params:   make(map[string][]byte),
		regulars: make(map[string]*regexp.Regexp),
		mode:     mode,
	}
	if qfe.mode == "" {
		qfe.mode = FilterModeAlwaysTrue
	}

	for k, v := range qfp {
		if k == ModeField {
			continue
		}
		qfe.params[k] = []byte(v)
		if mode == FilterModeRegular {
			if reg, err := regexp.Compile(v); err == nil {
				qfe.regulars[k] = reg
			} else {
				return nil, err
			}
		}
	}
	return qfe, nil
}

func (hfp HeaderFilterParams) To() (*HeaderFilterExecutor, error) {
	if hfp == nil {
		return &HeaderFilterExecutor{mode: FilterModeAlwaysTrue}, nil
	}

	mode := hfp[ModeField]
	hfe := &HeaderFilterExecutor{
		params:   make(map[string][]byte),
		regulars: make(map[string]*regexp.Regexp),
		mode:     mode,
	}
	if hfe.mode == "" {
		hfe.mode = FilterModeAlwaysTrue
	}

	for k, v := range hfp {
		if k == ModeField {
			continue
		}
		hfe.params[k] = []byte(v)
		if mode == FilterModeRegular {
			if reg, err := regexp.Compile(v); err == nil {
				hfe.regulars[k] = reg
			} else {
				return nil, err
			}
		}
	}
	return hfe, nil
}

func (bfp BodyFilterParams) To() (*BodyFilterExecutor, error) {
	if bfp == nil {
		return &BodyFilterExecutor{mode: FilterModeAlwaysTrue}, nil
	}

	mode := bfp[ModeField]
	bfe := &BodyFilterExecutor{mode: mode}
	if bfe.mode == "" {
		bfe.mode = FilterModeAlwaysTrue
	}

	for k, v := range bfp {
		if k == ModeField {
			continue
		}

		switch mode {
		case FilterModeKeyword:
			bfe.keyword = []byte(v)

		case FilterModeRegular:
			reg, err := regexp.Compile(v)
			if err != nil {
				return nil, err
			}
			bfe.regular = reg
		}
	}
	return bfe, nil
}

func (tmp *Template) To() (*TemplateExecutor, error) {
	te := &TemplateExecutor{
		IsGolangTemplate: tmp.IsTemplate,
		IsBinData:        false,
		template:         nil,
	}

	if tmp.B64EncodedBody != "" {
		te.IsBinData = true
		body, err := base64.StdEncoding.DecodeString(tmp.B64EncodedBody)
		if err != nil {
			return nil, err
		}
		te.body = body
	} else {
		te.body = []byte(tmp.Body)
	}

	header := new(fasthttp.ResponseHeader)
	header.SetStatusCode(tmp.StatusCode)
	for k, v := range tmp.Header {
		header.Set(k, v)
	}
	te.header = header

	if te.IsGolangTemplate {
		tmpl, err := template.New(misc.GenRandomString(8)).Funcs(defaultTemplateFuncs).Parse(string(te.body))
		if err != nil {
			return nil, err
		}
		te.template = tmpl
	}
	return te, nil
}
