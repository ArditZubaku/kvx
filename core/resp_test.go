package core_test

import (
	"errors"
	"fmt"
	"math"
	"strconv"
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

func TestReadInt64_Overflow(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "Max Int64",
			input:   ":" + strconv.FormatInt(math.MaxInt64, 10) + "\r\n",
			wantErr: nil,
		},
		{
			name:    "Overflow Max Int64 by 1",
			input:   ":9223372036854775808\r\n", // Max + 1
			wantErr: core.ErrIntegerOverflow,
		},
		{
			name:    "Massive Overflow",
			input:   ":9999999999999999999999999\r\n",
			wantErr: core.ErrIntegerOverflow,
		},
		{
			name:    "Min Int64",
			input:   ":" + strconv.FormatInt(math.MinInt64, 10) + "\r\n",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := core.Decode([]byte(tt.input))
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("readInt64() error = %v, wantErr %v, input %v", err, tt.wantErr, tt.input)
			}
		})
	}
}

func TestBulkStringDecode(t *testing.T) {
	cases := map[string]string{
		"$5\r\nhello\r\n": "hello",
		"$0\r\n\r\n":      "",
	}

	// TODO: I should add tests for the failing cases as well
	execCases(t, cases)
}

func TestArrayDecode(t *testing.T) {
	cases := map[string][]any{
		"*0\r\n":                                                   {},
		"*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n":                     {"hello", "world"},
		"*3\r\n:1\r\n:2\r\n:3\r\n":                                 {int64(1), int64(2), int64(3)},
		"*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$5\r\nhello\r\n":            {int64(1), int64(2), int64(3), int64(4), "hello"},
		"*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Hello\r\n-World\r\n": {[]int64{int64(1), int64(2), int64(3)}, []any{"Hello", "World"}},
	}

	for k, v := range cases {
		arr, _ := core.Decode([]byte(k))
		t.Log("arr", arr)
		if len(arr) > 0 {
			arr = arr[0].([]any)
		}

		if len(arr) != len(v) {
			t.Fail()
		}
		t.Log("arr", len(arr))
		t.Log("v", len(v))
		for i := range arr {
			if fmt.Sprintf("%v", v[i]) != fmt.Sprintf("%v", arr[i]) {
				t.Fail()
			}
		}
	}
}

func execCases[T comparable](t *testing.T, cases map[string]T) {
	for k, v := range cases {
		res, _ := core.Decode([]byte(k))
		if res[0] != v {
			t.Fail()
		}
	}
}
