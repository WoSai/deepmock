package client

import (
	"fmt"

	"github.com/jacexh/requests"
	"github.com/wosai/deepmock/types"
)

type (
	// DeepMockClient Client对象
	DeepMockClient struct {
		url    string
		client *requests.Session
	}

	// DeepMockError client error结构体
	DeepMockError struct {
		code int
		err  string
	}

	// Response client通用的返回响应
	Response struct {
		Code         int    `json:"code"`
		ErrorMessage string `json:"err_msg,omitempty"`
	}

	// RuleResponse 单规则接口返回报文
	RuleResponse struct {
		Response
		Data *types.RuleDO `json:"data,omitempty"`
	}

	// RulesResponse 多规则接口返回报文
	RulesResponse struct {
		Response
		Data []*types.RuleDO `json:"data,omitempty"`
	}
)

const (
	entrypointRule  = "/api/v1/rule"
	entrypointRules = "/api/v1/rules"

	returnCodeOK = 200
)

// NewDeepMockError client error工厂函数
func NewDeepMockError(res Response) *DeepMockError {
	return &DeepMockError{
		code: res.Code,
		err:  res.ErrorMessage,
	}
}

// Error error的实现
func (e *DeepMockError) Error() string {
	return fmt.Sprintf("[%d]: %s", e.code, e.err)
}

// NewDeepMockClient client的工厂函数
func NewDeepMockClient(url string) *DeepMockClient {
	session := requests.NewSession(requests.Option{Name: "DeepMock Go Client"})
	return &DeepMockClient{
		url:    url,
		client: session,
	}
}

// CreateMockRule 创建规则接口
func (c *DeepMockClient) CreateMockRule(rule *types.RuleDO) (*types.RuleDO, error) {
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

// DeleteMockRule 删除规则接口
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

// GetMockRule 获取规则接口
func (c *DeepMockClient) GetMockRule(rid string) (*types.RuleDO, error) {
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

// PutMockRule 全量更新规则接口
func (c *DeepMockClient) PutMockRule(rule *types.RuleDO) (*types.RuleDO, error) {
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

// PatchMockRule 部分更新规则接口
func (c *DeepMockClient) PatchMockRule(rule *types.RuleDO) (*types.RuleDO, error) {
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

// ExportRules 导出所有规则
func (c *DeepMockClient) ExportRules() ([]*types.RuleDO, error) {
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

// ImportRules 导入规则
func (c *DeepMockClient) ImportRules(rules ...*types.RuleDO) error {
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
