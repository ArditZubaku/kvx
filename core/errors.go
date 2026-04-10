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
)
