package core_test

import (
	"testing"

	"github.com/ArditZubaku/kvx/core"
)

func TestSimpleStringDecode(t *testing.T) {
	cases := map[string]string{
		"+OK\r\n": "OK",
	}
	for k, v := range cases {
		res, _ := core.Decode([]byte(k))
		if res != v {
			t.Fail()
		}
	}
}

func TestError(t *testing.T) {
}
