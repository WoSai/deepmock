package deepmock

import (
	"github.com/qastub/deepmock/types"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

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

func HandleCreateRule(ctx *fasthttp.RequestCtx, next func(error)) {
	rule := new(types.ResourceRule)
	if err := bindBody(ctx, rule); err != nil {
		return
	}

	re, err := defaultRuleManager.createRule(rule)
	res := new(types.CommonResource)
	if err != nil {
		res.Code = fasthttp.StatusBadRequest
		res.ErrorMessage = err.Error()
	} else {
		res.Code = 200
		res.Data = re.wrap()
	}

	data, _ := json.Marshal(res)
	ctx.Response.Header.SetContentType("application/json")
	ctx.Response.SetBody(data)
}

func HandleGetRule(ctx *fasthttp.RequestCtx, next func(error)) {

}

func HandleDeleteRule(ctx *fasthttp.RequestCtx, next func(error)) {

}

func HandleUpdateRule(ctx *fasthttp.RequestCtx, next func(error)) {

}

func HandlePatchRule(ctx *fasthttp.RequestCtx, next func(error)) {

}

func HandleExportRules(ctx *fasthttp.RequestCtx, next func(error)) {

}

func HandleImportRules(ctx *fasthttp.RequestCtx, next func(error)) {

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
