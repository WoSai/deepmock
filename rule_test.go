package deepmock

import (
	"html/template"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestMatch_Match(t *testing.T) {
	rm, err := newRequestMatcher(&ResourceRequestMatcher{Path: "/", Method: "Get"})
	assert.Nil(t, err)
	assert.True(t, rm.match([]byte("/"), []byte("GET")))

	rm, err = newRequestMatcher(&ResourceRequestMatcher{Method: "GET", Path: "/api/v1/create"})
	assert.Nil(t, err)
	assert.False(t, rm.match([]byte("/api/v1/create"), []byte("POST")))

	rm, err = newRequestMatcher(&ResourceRequestMatcher{Method: "GET", Path: "/api/v1/create"})
	assert.Nil(t, err)
	assert.False(t, rm.match([]byte("/api/v1/update"), []byte("GET")))

	rm, err = newRequestMatcher(&ResourceRequestMatcher{Method: "GET", Path: "/api/v[0-9]+/create"})
	assert.Nil(t, err)
	assert.True(t, rm.match([]byte("/api/v10/create"), []byte("GET")))
	assert.False(t, rm.match([]byte("/api/va/create"), []byte("GET")))
}

func TestHTMLTemplate(t *testing.T) {
	type Inventory struct {
		Material string
		Count    uint
	}
	sweaters := Inventory{"wool", 17}
	tmpl, err := template.New("test").Parse("{{.Count}} items are made of {{.Material}}\n")
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, sweaters)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(os.Stdout, sweaters)
	if err != nil {
		panic(err)
	}
}
