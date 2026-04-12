package core

import (
	"io"
	"strconv"
)

var okResponse = []byte("+OK\r\n")

func EvalAndRespond(cmd *RedisCmd, conn io.ReadWriter) error {
	switch cmd.Cmd {
	case "PING":
		return evalPING(cmd.Args, conn)
	case "SET":
		return evalSET(cmd.Args, conn)
	case "GET":
		return evalGET(cmd.Args, conn)
	case "TTL":
		return evalTTL(cmd.Args, conn)
	default:
		// for now
		return evalPING(cmd.Args, conn)
	}
}

func evalPING(args []string, conn io.ReadWriter) error {
	var b []byte

	if len(args) >= 2 {
		return ErrPingInvalidArgs
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	// TODO: Rethink whether this should write the data or just return it
	_, err := conn.Write(b)
	return err
}

// evalSET - SET key value [NX | XX] [GET] [EX seconds | PX milliseconds | EXAT unix-sec | PXAT unix-ms] [KEEPTTL]
func evalSET(args []string, conn io.ReadWriter) error {
	if len(args) <= 1 {
		return ErrSetInvalidArgs
	}

	var expirationMs int64 = -1 // default value - never expire

	key, value := args[0], args[1]

	// since key and value are mandatory, we start from the 3rd arg,
	// everything from here is optional
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++ // move one step, there should be the expiration value
			if i == len(args) {
				return ErrSyntaxError
			}

			// expirationSec, err := strconv.ParseInt(args[3], 10, 64)
			expirationSec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return ErrNotIntegerOutOfRange
			}

			// we get the value in seconds but we should store it in ms
			expirationMs = expirationSec * 1_000
		default:
			return ErrSyntaxError
		}
	}

	// put the key and value in a hash table
	Put(key, NewObj(value, expirationMs))

	conn.Write(okResponse)

	return nil
}

func evalGET(args []string, conn io.ReadWriter) error {
	panic("unimplemented")
}

func evalTTL(args []string, conn io.ReadWriter) error {
	panic("unimplemented")
}
