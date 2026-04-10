package core_test

import (
	"testing"

	"github.com/ArditZubaku/kvx/core"
)

func TestSimpleStringDecode(t *testing.T) {
	cases := map[string]string{
		"+OK\r\n": "OK",
	}

	execCases(t, cases)
}

func TestError(t *testing.T) {
	cases := map[string]string{
		"-Error message\r\n": "Error message",
	}

	execCases(t, cases)
}

func execCases(t *testing.T, cases map[string]string) {
	for k, v := range cases {
		res, _ := core.Decode([]byte(k))
		if res != v {
			t.Fail()
		}
	}
}
