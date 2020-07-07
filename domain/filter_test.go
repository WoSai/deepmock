package domain

//func TestEmptyRequestFilter(t *testing.T) {
//	req := fasthttp.AcquireRequest()
//	defer fasthttp.ReleaseRequest(req)
//
//	req.Header.SetMethod("POST")
//	req.SetRequestURI("/api/v1/query")
//	req.URI().SetQueryString("start=2019-09-01&end=2019-09-02")
//	req.Header.SetContentType("application/json; charset=UTF-8")
//	req.Header.Set("X-Version", "1.0")
//	req.SetBody([]byte(`{"hello":"deepmock"}`))
//
//	rf := new(requestFilter)
//	assert.True(t, rf.filter(req))
//
//	rf = &requestFilter{}
//	assert.True(t, rf.filter(req))
//}
