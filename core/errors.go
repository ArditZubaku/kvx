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

	// This is the exact error message Redis returns
	ErrPingInvalidArgs      = errors.New("ERR wrong number of arguments for 'ping' command")
	ErrSetInvalidArgs       = errors.New("(error) ERR wrong number of arguments for 'set' command")
	ErrSyntaxError          = errors.New("(error) ERR syntax error")
	ErrNotIntegerOutOfRange = errors.New("(error) ERR value is not an integer or out of range")
)
