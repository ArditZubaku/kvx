package server

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/ArditZubaku/kvx/core"
)

func readCommands(conn io.ReadWriter) ([]*core.RedisCmd, error) {
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, io.EOF
	}

	values, err := core.Decode(buf[:n])
	if err != nil {
		return nil, err
	}

	// TODO: Preallocate
	// cmds := make([]*core.RedisCmd, len(values))
	cmds := make([]*core.RedisCmd, 0)
	for _, val := range values {
		tokens, err := toArrayString(val.([]any))
		if err != nil {
			return nil, err
		}

		cmds = append(
			cmds,
			&core.RedisCmd{
				Cmd:  strings.ToUpper(tokens[0]),
				Args: tokens[1:],
			},
		)
	}

	return cmds, nil
}

func toArrayString(arr []any) ([]string, error) {
	strArr := make([]string, len(arr))

	for i := range arr {
		val, ok := arr[i].(string)
		if !ok {
			return nil, core.ErrInvalidType
		}
		strArr[i] = val
	}

	return strArr, nil
}

func respond(cmd []*core.RedisCmd, c *core.Client) {
	if err := core.EvalAndRespond(cmd, c); err != nil {
		// respond with error
		_, fErr := fmt.Fprintf(c, "-%s\r\n", err)
		if fErr != nil {
			log.Printf("failed to write response: %v\n", err)
			return
		}
	}
}
