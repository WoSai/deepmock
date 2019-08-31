package deepmock

import (
	"bytes"
	"errors"

	"github.com/qastub/deepmock/types"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

var (
	slash          = []byte(`/`)
	apiGetRulePath = []byte(`/api/v1/rule`)
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
func HandleMockedAPI(ctx *fasthttp.RequestCtx, next func(error)) {
	re, founded := defaultRuleManager.findExecutor(ctx.Request.URI().Path(), ctx.Request.Header.Method())
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
		c := re.context
		w := re.weightPicker.dice()
		q := extractQueryAsParams(&ctx.Request)
		f, j := extractBodyAsParams(&ctx.Request)

		rc := renderContext{Context: c, Weight: w, Query: q, Form: f, Json: j}
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
func HandleCreateRule(ctx *fasthttp.RequestCtx, next func(error)) {
	rule := new(types.ResourceRule)
	if err := bindBody(ctx, rule); err != nil {
		return
	}

	re, err := defaultRuleManager.createRule(rule)
	renderAPIResponse(&ctx.Response, re, err)
}

// HandleGetRule 根据rule id获取规则
func HandleGetRule(ctx *fasthttp.RequestCtx, next func(error)) {
	ruleID := parsePathVar(apiGetRulePath, ctx.RequestURI())

	re, exists := defaultRuleManager.getRuleByID(ruleID)
	var err error
	if !exists {
		err = errors.New("cannot found rule with id " + ruleID)
	}
	renderAPIResponse(&ctx.Response, re, err)
}

// HandleDeleteRule 根据rule id删除规则
func HandleDeleteRule(ctx *fasthttp.RequestCtx, next func(error)) {
	res := new(types.ResourceRule)
	if err := bindBody(ctx, res); err != nil {
		return
	}

	defaultRuleManager.deleteRule(res)
	renderAPIResponse(&ctx.Response, nil, nil)
}

// HandleUpdateRule 根据rule id更新目前规则，如果规则不存在，不会新建
func HandleUpdateRule(ctx *fasthttp.RequestCtx, next func(error)) {
	res := new(types.ResourceRule)
	if err := bindBody(ctx, res); err != nil {
		return
	}

	re, err := defaultRuleManager.updateRule(res)
	renderAPIResponse(&ctx.Response, re, err)
}

// HandlePatchRule 根据rule id更新目前规则，与put的区别在于：put需要传入完整的rule对象，而patch只需要传入更新部分即可
func HandlePatchRule(ctx *fasthttp.RequestCtx, next func(error)) {
	res := new(types.ResourceRule)
	if err := bindBody(ctx, res); err != nil {
		return
	}

	re, err := defaultRuleManager.patchRule(res)
	renderAPIResponse(&ctx.Response, re, err)
}

// HandleExportRules 导出当前所有规则
func HandleExportRules(ctx *fasthttp.RequestCtx, next func(error)) {
	re := defaultRuleManager.exportRules()
	rules := make([]*types.ResourceRule, len(re))
	for k, v := range re {
		rules[k] = v.wrap()
	}

	renderAPIResponse(&ctx.Response, rules, nil)
}

// HandleImportRules 导入规则，将会清空目前所有规则
func HandleImportRules(ctx *fasthttp.RequestCtx, next func(error)) {
	var rules []*types.ResourceRule
	if err := bindBody(ctx, &rules); err != nil {
		return
	}
	err := defaultRuleManager.importRules(rules...)
	renderAPIResponse(&ctx.Response, nil, err)
}

func HandleHelp(ctx *fasthttp.RequestCtx, next func(error)) {

}

func bindBody(ctx *fasthttp.RequestCtx, v interface{}) error {
	if err := json.Unmarshal(ctx.Request.Body(), v); err != nil {
		Logger.Error("failed to parse request body", zap.ByteString("path", ctx.Request.URI().Path()), zap.ByteString("method", ctx.Request.Header.Method()), zap.Error(err))
		ctx.Response.Header.SetContentType("application/json")
		ctx.Response.SetStatusCode(fasthttp.StatusOK)

		res := new(types.CommonResource)
		res.Code = fasthttp.StatusBadRequest
		res.ErrorMessage = err.Error()
		data, _ := json.Marshal(res)
		ctx.SetBody(data)
		return err
	}
	return nil
}

func renderAPIResponse(resp *fasthttp.Response, v interface{}, err error) {
	res := new(types.CommonResource)
	if err == nil {
		res.Code = 200
		res.Data = v
	} else {
		res.Code = 400
		res.ErrorMessage = err.Error()
	}

	data, _ := json.Marshal(res)
	resp.Header.SetContentType("application/json")
	resp.SetBody(data)
}
