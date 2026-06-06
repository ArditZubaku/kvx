package core

import "errors"

var (
	ErrNoData              = errors.New("no data")
	ErrInvalidSimpleString = errors.New("invalid simple string")
	ErrInvalidInt64        = errors.New("invalid int64")
	ErrInvalidArray        = errors.New("invalid array")
	ErrInvalidBulkString   = errors.New("invalid bulk string")
	ErrMissingCRLF         = errors.New("invalid simple string: missing CRLF")
	ErrIntegerOverflow     = errors.New("integer overflow")
	ErrInvalidType         = errors.New("invalid type")
	ErrUnknownDataType     = errors.New("unknown RESP data type")

	// This is the exact error message Redis returns
	ErrPingInvalidArgs      = errors.New("ERR wrong number of arguments for 'ping' command")
	ErrSetInvalidArgs       = errors.New("ERR wrong number of arguments for 'set' command")
	ErrGetInvalidArgs       = errors.New("ERR wrong number of arguments for 'get' command")
	ErrTTLInvalidArgs       = errors.New("ERR wrong number of arguments for 'ttl' command")
	ErrExpireInvalidArgs    = errors.New("ERR wrong number of arguments for 'expire' command")
	ErrSyntaxError          = errors.New("ERR syntax error")
	ErrNotIntegerOutOfRange = errors.New("ERR value is not an integer or out of range")
	ErrINCRInvalidArgs      = errors.New("ERR wrong number of arguments for 'incr' command")
)
