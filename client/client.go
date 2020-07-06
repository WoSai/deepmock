package client

import (
	"fmt"

	"github.com/jacexh/requests"
	"github.com/wosai/deepmock/types"
)

type (
	DeepMockClient struct {
		url    string
		client *requests.Session
	}

	DeepMockError struct {
		code int
		err  string
	}

	Response struct {
		Code         int    `json:"code"`
		ErrorMessage string `json:"err_msg,omitempty"`
	}

	RuleResponse struct {
		Response
		Data *types.RuleDo `json:"data,omitempty"`
	}

	RulesResponse struct {
		Response
		Data []*types.RuleDo `json:"data,omitempty"`
	}
)

const (
	entrypointRule  = "/api/v1/rule"
	entrypointRules = "/api/v1/rules"

	returnCodeOK = 200
)

func NewDeepMockError(res Response) *DeepMockError {
	return &DeepMockError{
		code: res.Code,
		err:  res.ErrorMessage,
	}
}

func (e *DeepMockError) Error() string {
	return fmt.Sprintf("[%d]: %s", e.code, e.err)
}

func NewDeepMockClient(url string) *DeepMockClient {
	session := requests.NewSession(requests.Option{Name: "DeepMock Go Client"})
	return &DeepMockClient{
		url:    url,
		client: session,
	}
}

func (c *DeepMockClient) CreateMockRule(rule *types.RuleDo) (*types.RuleDo, error) {
	res := new(RuleResponse)
	_, _, err := c.client.Post(c.url+entrypointRule, requests.Params{Json: rule}, requests.UnmarshalJSONResponse(res))
	if err != nil {
		return nil, err
	}
	if res.Code != returnCodeOK {
		return nil, NewDeepMockError(res.Response)
	}
	return res.Data, nil
}

func (c *DeepMockClient) DeleteMockRule(rid string) error {
	res := new(RuleResponse)
	_, _, err := c.client.Delete(c.url+entrypointRule, requests.Params{Json: requests.Any{"id": rid}}, requests.UnmarshalJSONResponse(res))
	if err != nil {
		return err
	}
	if res.Code != returnCodeOK {
		return NewDeepMockError(res.Response)
	}
	return nil
}

func (c *DeepMockClient) GetMockRule(rid string) (*types.RuleDo, error) {
	res := new(RuleResponse)
	_, _, err := c.client.Get(c.url+entrypointRule+"/"+rid, requests.Params{}, requests.UnmarshalJSONResponse(res))
	if err != nil {
		return nil, err
	}
	if res.Code != returnCodeOK {
		return nil, NewDeepMockError(res.Response)
	}
	return res.Data, nil
}

func (c *DeepMockClient) PutMockRule(rule *types.RuleDo) (*types.RuleDo, error) {
	res := new(RuleResponse)
	_, _, err := c.client.Put(c.url+entrypointRule, requests.Params{Json: rule}, requests.UnmarshalJSONResponse(res))
	if err != nil {
		return nil, err
	}
	if res.Code != returnCodeOK {
		return nil, NewDeepMockError(res.Response)
	}
	return res.Data, nil
}

func (c *DeepMockClient) PatchMockRule(rule *types.RuleDo) (*types.RuleDo, error) {
	res := new(RuleResponse)
	_, _, err := c.client.Patch(c.url+entrypointRule, requests.Params{Json: rule}, requests.UnmarshalJSONResponse(res))
	if err != nil {
		return nil, err
	}
	if res.Code != returnCodeOK {
		return nil, NewDeepMockError(res.Response)
	}
	return res.Data, nil
}

func (c *DeepMockClient) ExportRules() ([]*types.RuleDo, error) {
	res := new(RulesResponse)
	_, _, err := c.client.Get(c.url+entrypointRules, requests.Params{}, requests.UnmarshalJSONResponse(res))
	if err != nil {
		return nil, err
	}
	if res.Code != returnCodeOK {
		return nil, NewDeepMockError(res.Response)
	}
	return res.Data, nil
}

func (c *DeepMockClient) ImportRules(rules ...*types.RuleDo) error {
	res := new(RulesResponse)
	_, _, err := c.client.Post(c.url+entrypointRules, requests.Params{Json: rules}, requests.UnmarshalJSONResponse(res))
	if err != nil {
		return err
	}
	if res.Code != returnCodeOK {
		return NewDeepMockError(res.Response)
	}
	return nil
}
