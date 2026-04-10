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

func TestInt64(t *testing.T) {
	cases := map[string]int64{
		":0\r\n":    0,
		":1000\r\n": 1000,
		":-1\r\n":   -1,
	}

	execCases(t, cases)
}

func execCases[T comparable](t *testing.T, cases map[string]T) {
	for k, v := range cases {
		res, _ := core.Decode([]byte(k))
		if res != v {
			t.Fail()
		}
	}
}
