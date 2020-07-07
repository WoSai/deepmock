package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"github.com/wosai/deepmock/application"
	"github.com/wosai/deepmock/misc"
	"github.com/wosai/deepmock/types"
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
	err := application.MockApplication.MockAPI(ctx)
	if err != nil {
		if errors.Is(err, application.ErrRuleNotFound) {
			misc.Logger.Error("no rule match your request")
		}
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
}

// HandleCreateRule 创建规则接口
func HandleCreateRule(ctx *fasthttp.RequestCtx, _ func(error)) {
	rule := new(types.RuleDTO)
	if err := bindBody(ctx, rule); err != nil {
		return
	}

	rid, err := application.MockApplication.CreateRule(context.TODO(), rule)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	rule, err = application.MockApplication.GetRule(context.TODO(), rid)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, rule)
}

// HandleGetRule 根据rule id获取规则
func HandleGetRule(ctx *fasthttp.RequestCtx, _ func(error)) {
	ruleID := parsePathVar(apiGetRulePath, ctx.RequestURI())

	rule, err := application.MockApplication.GetRule(context.TODO(), ruleID)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, rule)
}

// HandleDeleteRule 根据rule id删除规则
func HandleDeleteRule(ctx *fasthttp.RequestCtx, _ func(error)) {
	res := new(types.RuleDTO)
	if err := bindBody(ctx, res); err != nil {
		return
	}

	err := application.MockApplication.DeleteRule(context.TODO(), res.ID)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, nil)
}

// HandlePutRule 根据rule id更新目前规则，如果规则不存在，不会新建
func HandlePutRule(ctx *fasthttp.RequestCtx, _ func(error)) {
	res := new(types.RuleDTO)
	if err := bindBody(ctx, res); err != nil {
		return
	}

	err := application.MockApplication.PutRule(context.TODO(), res)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	rule, err := application.MockApplication.GetRule(context.TODO(), res.ID)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, rule)
}

// HandlePatchRule 根据rule id更新目前规则，与put的区别在于：put需要传入完整的rule对象，而patch只需要传入更新部分即可
func HandlePatchRule(ctx *fasthttp.RequestCtx, _ func(error)) {
	res := new(types.RuleDTO)
	if err := bindBody(ctx, res); err != nil {
		return
	}

	err := application.MockApplication.PatchRule(context.TODO(), res)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	rule, err := application.MockApplication.GetRule(context.TODO(), res.ID)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, rule)
}

// HandleExportRules 导出当前所有规则
func HandleExportRules(ctx *fasthttp.RequestCtx, _ func(error)) {
	rules, err := application.MockApplication.Export(context.TODO())
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, rules)
}

// HandleImportRules 导入规则，将会清空目前所有规则
func HandleImportRules(ctx *fasthttp.RequestCtx, _ func(error)) {
	var rules []*types.RuleDTO
	if err := bindBody(ctx, &rules); err != nil {
		return
	}

	err := application.MockApplication.Import(context.TODO(), rules...)
	if err != nil {
		renderFailedAPIResponse(&ctx.Response, err)
		return
	}
	renderSuccessfulResponse(&ctx.Response, nil)
}

func bindBody(ctx *fasthttp.RequestCtx, v interface{}) error {
	if err := json.Unmarshal(ctx.Request.Body(), v); err != nil {
		misc.Logger.Error("failed to parse request body", zap.ByteString("path", ctx.Request.URI().Path()), zap.ByteString("method", ctx.Request.Header.Method()), zap.Error(err))
		ctx.Response.Header.SetContentType("application/json")
		ctx.Response.SetStatusCode(fasthttp.StatusOK)

		res := new(types.CommonResponseDTO)
		res.Code = fasthttp.StatusBadRequest
		res.ErrorMessage = err.Error()
		data, _ := json.Marshal(res)
		ctx.SetBody(data)
		return err
	}
	return nil
}

func renderSuccessfulResponse(resp *fasthttp.Response, v interface{}) {
	res := &types.CommonResponseDTO{
		Code: http.StatusOK,
		Data: v,
	}
	data, _ := json.Marshal(res)
	resp.Header.SetContentType("application/json")
	resp.SetBody(data)
}

func renderFailedAPIResponse(resp *fasthttp.Response, err error) {
	res := &types.CommonResponseDTO{Code: http.StatusBadRequest, ErrorMessage: err.Error()}
	data, _ := json.Marshal(res)
	resp.Header.SetContentType("application/json")
	resp.SetBody(data)
}
