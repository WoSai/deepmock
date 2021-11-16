package misc

import (
	"fmt"
	"testing"
)

func TestGenID(t *testing.T) {
	fmt.Println(GenID([]byte("/rpc/token"), []byte("POST")))
	fmt.Println(GenID([]byte("^/$"), []byte("GET")))
}
