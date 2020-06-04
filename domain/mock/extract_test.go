package mock

import (
	"mime/multipart"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestExtractFromHeader(t *testing.T) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.Set("Version", "123")
	req.Header.SetBytesV("X-Env-Flag", []byte("base"))

	p := extractHeaderAsParams(req)
	assert.EqualValues(t, p, params{"Version": "123", "X-Env-Flag": "base"})
}

func TestExtractFromQueryString(t *testing.T) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.URI().SetQueryString("name=foobar&message=欢迎")

	p := extractQueryAsParams(req)
	assert.EqualValues(t, p, params{"name": "foobar", "message": "欢迎"})
}

func TestExtractFromUrlencodedForm(t *testing.T) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("POST")
	data := url.Values{}
	data.Set("name", "foobar")
	data.Set("message", "中国")
	req.Header.SetContentType("application/x-www-form-urlencoded; charset=UTF-8")
	req.SetBodyString(data.Encode())

	f, _ := extractBodyAsParams(req)
	assert.EqualValues(t, f, params{"name": "foobar", "message": "中国"})

	args := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(args)
	args.Set("name", "foobar")
	args.Set("message", "中国")
	req.SetBody(args.QueryString())

	f, _ = extractBodyAsParams(req)
	assert.EqualValues(t, f, params{"name": "foobar", "message": "中国"})
}

func TestExtractFromMultipartForm(t *testing.T) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("POST")
	writer := multipart.NewWriter(req.BodyWriter())
	req.Header.SetContentType(writer.FormDataContentType())
	assert.Nil(t, writer.WriteField("name", "foobar"))
	assert.Nil(t, writer.WriteField("message", "中国"))
	assert.Nil(t, writer.Close())

	f, _ := extractBodyAsParams(req)
	assert.EqualValues(t, params{"name": "foobar", "message": "中国"}, f)
}

func TestExtractFromJson(t *testing.T) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetContentType("application/json; charset=UTF-8")
	req.SetBody([]byte(`{"name":"foobar", "message":"中国"}`))
	req.Header.SetMethod("POST")

	_, j := extractBodyAsParams(req)
	assert.EqualValues(t, map[string]interface{}{"name": "foobar", "message": "中国"}, j)
}

func TestExtractUnsupportedContentType(t *testing.T) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("image/png")
	req.SetBody([]byte(`{"name":"foobar"}`))
	f, j := extractBodyAsParams(req)
	assert.Nil(t, f)
	assert.Nil(t, j)
}
