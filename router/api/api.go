package api

import (
	"bytes"
	"errors"

	"github.com/wosai/deepmock/service"

	jsoniter "github.com/json-iterator/go"

	"github.com/valyala/fasthttp"
	"github.com/wosai/deepmock"
	"github.com/wosai/deepmock/types"
	"github.com/wosai/deepmock/types/resource"
	"go.uber.org/zap"
)

var (
	slash          = []byte(`/`)
	apiGetRulePath = []byte(`/api/v1/rule`)
	json           = jsoniter.ConfigCompatibleWithStandardLibrary
)

func parsePathVar(path, uri []byte) string {
	if bytes.Compare(path, uri) == 1 {
		panic(errors.New("bad request uir"))
	}

	external := bytes.Split(uri[len(path)-1:], slash) // 忽略第一个 / 符号
	if len(external) >= 2 {
		return string(external[1])
	}
	return ""
}

// HandleMockedAPI 处理所有mock api
func HandleMockedAPI(ctx *fasthttp.RequestCtx, _ func(error)) {
	re, founded, _ := defaultRuleManager.findExecutor(ctx.Request.URI().Path(), ctx.Request.Header.Method())
	if !founded {
		res := new(types.CommonResource)
		res.Code = 400
		res.ErrorMessage = "no rule match your request"
		data, _ := json.Marshal(res)
		ctx.Response.Header.SetContentType("application/json")
		ctx.Response.SetBody(data)
		return
	}

	regulation := re.visitBy(&ctx.Request)

	// 没有任何模板匹配到
	if regulation == nil {
		res := new(types.CommonResource)
		res.Code = 400
		res.ErrorMessage = "missing matched response regulation"
		data, _ := json.Marshal(res)
		ctx.Response.Header.SetContentType("application/json")
		ctx.Response.SetBody(data)
		return
	}
	render(re, regulation.responseTemplate, ctx)
}

func render(re *ruleExecutor, rt *responseTemplate, ctx *fasthttp.RequestCtx) {
	rt.header.CopyTo(&ctx.Response.Header)
	if rt.isTemplate {
		c := re.variable
		w := re.weightPicker.dice()
		h := extractHeaderAsParams(&ctx.Request)
		q := extractQueryAsParams(&ctx.Request)
		f, j := extractBodyAsParams(&ctx.Request)

		rc := renderContext{Variable: c, Weight: w, Header: h, Query: q, Form: f, Json: j}
		if err := rt.renderTemplate(rc, ctx.Response.BodyWriter()); err != nil {
			Logger.Error("failed to render response template", zap.Error(err))
			res := new(types.CommonResource)
			res.Code = fasthttp.StatusBadRequest
			res.ErrorMessage = err.Error()
			data, _ := json.Marshal(res)
			ctx.Response.SetBody(data)
			return
		}
		return
	}

	ctx.Response.SetBody(rt.body)
}

// HandleCreateRule 创建规则接口
func HandleCreateRule(ctx *fasthttp.RequestCtx, _ func(error)) {
	rule := new(resource.Rule)
	if err := bindBody(ctx, rule); err != nil {
		return
	}

	rid, err := service.RuleService.CreateRule(rule)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	rule, err = service.RuleService.GetRule(rid)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, rule)
}

// HandleGetRule 根据rule id获取规则
func HandleGetRule(ctx *fasthttp.RequestCtx, _ func(error)) {
	ruleID := parsePathVar(apiGetRulePath, ctx.RequestURI())

	rule, err := service.RuleService.GetRule(ruleID)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, rule)
}

// HandleDeleteRule 根据rule id删除规则
func HandleDeleteRule(ctx *fasthttp.RequestCtx, _ func(error)) {
	res := new(resource.Rule)
	if err := bindBody(ctx, res); err != nil {
		return
	}

	//defaultRuleManager.deleteRule(res)
	err := service.RuleService.DeleteRule(res.ID)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, nil)
}

// HandlePutRule 根据rule id更新目前规则，如果规则不存在，不会新建
func HandlePutRule(ctx *fasthttp.RequestCtx, _ func(error)) {
	res := new(resource.Rule)
	if err := bindBody(ctx, res); err != nil {
		return
	}

	err := service.RuleService.PutRule(res)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	rule, err := service.RuleService.GetRule(res.ID)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, rule)
}

// HandlePatchRule 根据rule id更新目前规则，与put的区别在于：put需要传入完整的rule对象，而patch只需要传入更新部分即可
func HandlePatchRule(ctx *fasthttp.RequestCtx, _ func(error)) {
	res := new(resource.Rule)
	if err := bindBody(ctx, res); err != nil {
		return
	}

	err := service.RuleService.PatchRule(res)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	rule, err := service.RuleService.GetRule(res.ID)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, rule)
}

// HandleExportRules 导出当前所有规则
func HandleExportRules(ctx *fasthttp.RequestCtx, _ func(error)) {
	//re := defaultRuleManager.exportRules()
	//rules := make([]*types.ResourceRule, len(re))
	//for k, v := range re {
	//	rules[k] = v.wrap()
	//}

	rules, err := storage.export()
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, rules)
}

// HandleImportRules 导入规则，将会清空目前所有规则
func HandleImportRules(ctx *fasthttp.RequestCtx, _ func(error)) {
	var rules []*types.ResourceRule
	if err := bindBody(ctx, &rules); err != nil {
		return
	}
	//err := defaultRuleManager.importRules(rules...)
	//if err != nil {
	//	renderFailedAPIResponse(&ctx.Response, err)
	//	return
	//}
	//renderSuccessfulResponse(&ctx.Response, nil)

	err := storage.importRules(rules...)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, nil)
}

func bindBody(ctx *fasthttp.RequestCtx, v interface{}) error {
	if err := json.Unmarshal(ctx.Request.Body(), v); err != nil {
		deepmock.Logger.Error("failed to parse request body", zap.ByteString("path", ctx.Request.URI().Path()), zap.ByteString("method", ctx.Request.Header.Method()), zap.Error(err))
		ctx.Response.Header.SetContentType("application/json")
		ctx.Response.SetStatusCode(fasthttp.StatusOK)

		res := new(resource.CommonResource)
		res.Code = fasthttp.StatusBadRequest
		res.ErrorMessage = err.Error()
		data, _ := json.Marshal(res)
		ctx.SetBody(data)
		return err
	}
	return nil
}

func renderSuccessfulResponse(resp *fasthttp.Response, v interface{}) {
	res := &types.CommonResource{
		Code: 200,
		Data: v,
	}

	data, _ := json.Marshal(res)
	resp.Header.SetContentType("application/json")
	resp.SetBody(data)
}

func renderFailedAPIResponse(resp *fasthttp.Response, err error) {
	res := &resource.CommonResource{Code: 400, ErrorMessage: err.Error()}
	data, _ := json.Marshal(res)
	resp.Header.SetContentType("application/json")
	resp.SetBody(data)
}
