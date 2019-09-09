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
	rule := new(types.ResourceRule)
	if err := bindBody(ctx, rule); err != nil {
		return
	}

	re, err := defaultRuleManager.createRule(rule)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, re.wrap())
}

// HandleGetRule 根据rule id获取规则
func HandleGetRule(ctx *fasthttp.RequestCtx, _ func(error)) {
	ruleID := parsePathVar(apiGetRulePath, ctx.RequestURI())

	re, exists := defaultRuleManager.getRuleByID(ruleID)
	var err error
	if !exists {
		err = errors.New("cannot found rule with id " + ruleID)
	}
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, re.wrap())
}

// HandleDeleteRule 根据rule id删除规则
func HandleDeleteRule(ctx *fasthttp.RequestCtx, _ func(error)) {
	res := new(types.ResourceRule)
	if err := bindBody(ctx, res); err != nil {
		return
	}

	defaultRuleManager.deleteRule(res)
	renderSuccessfulResponse(&ctx.Response, nil)
}

// HandleUpdateRule 根据rule id更新目前规则，如果规则不存在，不会新建
func HandleUpdateRule(ctx *fasthttp.RequestCtx, _ func(error)) {
	res := new(types.ResourceRule)
	if err := bindBody(ctx, res); err != nil {
		return
	}

	re, err := defaultRuleManager.updateRule(res)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, re.wrap())
}

// HandlePatchRule 根据rule id更新目前规则，与put的区别在于：put需要传入完整的rule对象，而patch只需要传入更新部分即可
func HandlePatchRule(ctx *fasthttp.RequestCtx, _ func(error)) {
	res := new(types.ResourceRule)
	if err := bindBody(ctx, res); err != nil {
		return
	}

	re, err := defaultRuleManager.patchRule(res)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, re.wrap())
}

// HandleExportRules 导出当前所有规则
func HandleExportRules(ctx *fasthttp.RequestCtx, _ func(error)) {
	re := defaultRuleManager.exportRules()
	rules := make([]*types.ResourceRule, len(re))
	for k, v := range re {
		rules[k] = v.wrap()
	}

	renderSuccessfulResponse(&ctx.Response, rules)
}

// HandleImportRules 导入规则，将会清空目前所有规则
func HandleImportRules(ctx *fasthttp.RequestCtx, _ func(error)) {
	var rules []*types.ResourceRule
	if err := bindBody(ctx, &rules); err != nil {
		return
	}
	err := defaultRuleManager.importRules(rules...)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, nil)
}

func HandleHelp(ctx *fasthttp.RequestCtx, _ func(error)) {

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
	res := &types.CommonResource{Code: 400, ErrorMessage: err.Error()}
	data, _ := json.Marshal(res)
	resp.Header.SetContentType("application/json")
	resp.SetBody(data)
}
