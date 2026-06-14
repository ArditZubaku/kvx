package core

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"
)

var (
	okResponse           = []byte("+OK\r\n")
	nilResponse          = []byte("$-1\r\n")
	notExistsResponse    = []byte(":-2\r\n")
	neverExpiresResponse = []byte(":-1\r\n")
	zero                 = []byte(":0\r\n")
	one                  = []byte(":1\r\n")
)

func EvalAndRespond(cmds []*RedisCmd, conn io.ReadWriter) error {
	// TODO: Pick something better for the capacity
	// NOTE: The slice should have a length of 0
	response := make([]byte, 0, len(cmds)*2)
	buf := bytes.NewBuffer(response)

	for _, cmd := range cmds {
		switch cmd.Cmd {
		case "PING":
			buf.Write(evalPING(cmd.Args))
		case "SET":
			buf.Write(evalSET(cmd.Args))
		case "GET":
			buf.Write(evalGET(cmd.Args))
		case "TTL":
			buf.Write(evalTTL(cmd.Args))
		case "DEL":
			buf.Write(evalDEL(cmd.Args))
		case "EXPIRE":
			buf.Write(evalEXPIRE(cmd.Args))
		case "BGREWRITEAOF":
			buf.Write(evalBGREWRITEAOF())
		case "INCR":
			buf.Write(evalINCR(cmd.Args))
		case "INFO":
			buf.Write(evalINFO())
		case "CLIENT":
			buf.Write(evalCLIENT(cmd.Args))
		case "LATENCY":
			buf.Write(evalLATENCY(cmd.Args))
		default:
			buf.Write(evalPING(cmd.Args))
		}
	}

	_, err := conn.Write(buf.Bytes())
	return err
}

func evalPING(args []string) []byte {
	if len(args) > 1 {
		return Encode(ErrPingInvalidArgs)
	}

	if len(args) == 1 {
		return Encode(args[0])
	}

	return EncodeSimple("PONG")
}

// evalSET - SET key value [NX | XX] [GET] [EX seconds | PX milliseconds | EXAT unix-sec | PXAT unix-ms] [KEEPTTL]
//
// SET assigns a value to a key, optionally with conditions and expiration.
func evalSET(args []string) []byte {
	if len(args) <= 1 {
		return Encode(ErrSetInvalidArgs)
	}

	var expirationMs int64 = -1 // default value - never expire

	key, value := args[0], args[1]
	objType, objEnc := deduceTypeEncoding(value)

	// since key and value are mandatory, we start from the 3rd arg,
	// everything from here is optional
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++ // move one step, there should be the expiration value
			if i == len(args) {
				return Encode(ErrSyntaxError)
			}

			expirationSec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return Encode(ErrNotIntegerOutOfRange)
			}

			// we get the value in seconds but we should store it in ms
			expirationMs = expirationSec * 1_000
		default:
			return Encode(ErrSyntaxError)
		}
	}

	// put the key and value in a hash table
	Put(key, NewObj(value, expirationMs, objType, objEnc))

	return okResponse
}

// evalGET - GET key (it has to be exactly 1 arg)
//
// GET returns the value stored at a key (or nil if it doesn’t exist)
func evalGET(args []string) []byte {
	if len(args) != 1 {
		return Encode(ErrGetInvalidArgs)
	}

	key := args[0]
	obj := Get(key)

	if obj == nil {
		return nilResponse
	}

	if hasExpired(obj) {
		return nilResponse
	}

	return Encode(obj.Value)
}

// evalTTL - TTL key
//
// TTL returns the remaining time-to-live of a key in seconds (-1 if no expiry, -2 if missing)
func evalTTL(args []string) []byte {
	// it has to be exactly 1 arg
	if len(args) != 1 {
		return Encode(ErrTTLInvalidArgs)
	}

	key := args[0]
	obj := Get(key)

	// if key does not exist, return RESP encoded -2
	// denoting that the key does not exist (that's how Redis responds)
	if obj == nil {
		return notExistsResponse
	}

	// if object exists, but no expiration is set on it then send `-1` (meaning never expires)
	exp, isExpirySet := getExpiry(obj)
	if !isExpirySet {
		return neverExpiresResponse
	}

	if hasExpired(obj) {
		return notExistsResponse
	}

	remainingExpirationMs := exp - uint64(time.Now().UnixMilli())
	return Encode(remainingExpirationMs / 1_000)
}

// evalDEL - DEL key [key ...]
//
// DEL removes one or more keys from Redis and returns how many were successfully deleted.
func evalDEL(args []string) []byte {
	deleted := 0

	for _, key := range args {
		if ok := Del(key); ok {
			deleted++
		}
	}

	return Encode(deleted)
}

// evalEXPIRE - EXPIRE key seconds
//
// EXPIRE sets a time-to-live (in seconds) on a key, after which it will automatically be deleted.
func evalEXPIRE(args []string) []byte {
	if len(args) <= 1 {
		return Encode(ErrExpireInvalidArgs)
	}

	key := args[0]
	expirationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(ErrNotIntegerOutOfRange)
	}

	obj := Get(key)

	// return 0 if the timeout was not set,
	// e.g. key doesn't exist or operation skipped due to the provided arguments
	if obj == nil {
		return zero
	}

	// NOTE: If I ever switch to value semantics, I should use Set() here instead of modifying the obj
	setExpiry(obj, expirationSec*1_000)

	// return 1 if the timeout was set
	return one
}

// TODO: Make it async by forking a new process OR probably just a separate goroutine
func evalBGREWRITEAOF() []byte {
	dumpAllAOF()

	return okResponse
}

// evalINCR - INCR key
//
// Increments the number stored at key by one.
// If the key does not exist, it is set to 0 before performing the operation.
// An error is returned if the key contains a value of the wrong type
// or contains a string that can not be represented as integer.
func evalINCR(args []string) []byte {
	if len(args) != 1 {
		return Encode(ErrINCRInvalidArgs)
	}

	key := args[0]
	obj := Get(key)

	if obj == nil {
		obj = NewObj("0", -1, ObjTypeString, ObjEncodingInt)
		Put(key, obj)
	}

	if err := assertType(obj.TypeEncoding, ObjTypeString); err != nil {
		return Encode(err)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingInt); err != nil {
		return Encode(err)
	}

	i, err := strconv.ParseInt(obj.Value.(string), 10, 64)
	if err != nil {
		return Encode(err)
	}

	i++
	obj.Value = strconv.FormatInt(i, 10)

	return Encode(i)
}

// evalINFO - INFO [section [section ...]]
//
// The INFO command returns information and statistics about the server in a format
// that is simple to parse by computers and easy to read by humans.
func evalINFO() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 256))

	buf.WriteString("# Keyspace\r\n")

	for i := range KeySpaceStat {
		// if _, err := fmt.Fprintf(buf, "db%d:keys=%d,expires=0,avg_ttl=0\r\n", i, len(store)); err != nil {
		if _, err := fmt.Fprintf(buf, "db%d:keys=%d,expires=0,avg_ttl=0\r\n", i, KeySpaceStat[i]["keys"]); err != nil {
			log.Printf("error writing stats: %+v\n", err)
			continue
		}
	}

	return Encode(buf.String())
}

func evalLATENCY(s []string) []byte {
	panic("unimplemented")
}

func evalCLIENT(s []string) []byte {
	panic("unimplemented")
}
