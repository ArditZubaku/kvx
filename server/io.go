package server

import (
	"fmt"
	"io"
	"strings"

	"github.com/ArditZubaku/kvx/core"
)

func readCommand(conn io.ReadWriter) (*core.RedisCmd, error) {
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	tokens, err := core.DecodeArrayString(buf[:n])
	if err != nil {
		return nil, err
	}

	return &core.RedisCmd{
		Cmd:  strings.ToUpper(tokens[0]),
		Args: tokens[1:],
	}, nil
}

func respond(cmd *core.RedisCmd, conn io.ReadWriter) {
	if err := core.EvalAndRespond(cmd, conn); err != nil {
		// respond with error
		fmt.Fprintf(conn, "-%s\r\n", err)
	}
}
