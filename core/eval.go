package core

import (
	"io"
	"strconv"
	"time"
)

var okResponse = []byte("+OK\r\n")
var nilResponse = []byte("$-1\r\n")
var notExistsResponse = []byte(":-2\r\n")
var neverExpiresResponse = []byte(":-1\r\n")

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
		b = EncodeSimple("PONG")
	} else {
		b = Encode(args[0])
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

	_, err := conn.Write(okResponse)
	return err
}

// evalGET - GET key
func evalGET(args []string, conn io.ReadWriter) error {
	// it has to be exactly 1 arg
	if len(args) != 1 {
		return ErrGetInvalidArgs
	}

	key := args[0]
	obj := Get(key)

	// if key does not exist, return RESP encoded nil
	if obj == nil {
		_, err := conn.Write(nilResponse)
		return err
	}

	// if key already expired then return nil
	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		_, err := conn.Write(nilResponse)
		return err
	}

	// return the RESP encoded value
	_, err := conn.Write(Encode(obj.Value))
	return err
}

// evalTTL - TTL key
func evalTTL(args []string, conn io.ReadWriter) error {
	// it has to be exactly 1 arg
	if len(args) != 1 {
		return ErrTTLInvalidArgs
	}

	key := args[0]
	obj := Get(key)

	// if key does not exist, return RESP encoded -2
	// denoting that the key does not exist (that's how Redis responds)
	if obj == nil {
		_, err := conn.Write(notExistsResponse)
		return err
	}

	// if object exists, but no expiration is set on it then send `-1` (meaning never expires)
	if obj.ExpiresAt == -1 {
		_, err := conn.Write(neverExpiresResponse)
		return err
	}

	// compute the time remaining for the key to expire
	// return the RESP encoded value of it
	expirationMs := obj.ExpiresAt - time.Now().UnixMilli()

	// if key expired -> therefore key does not exist, return -2
	if expirationMs < 0 {
		_, err := conn.Write(notExistsResponse)
		return err
	}

	_, err := conn.Write(Encode(expirationMs / 1_000))
	return err
}
