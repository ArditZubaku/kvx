package core

import "errors"

var (
	ErrNoData              = errors.New("no data")
	ErrInvalidSimpleString = errors.New("invalid simple string")
	ErrInvalidInt64        = errors.New("invalid int64")
	ErrMissingCRLF         = errors.New("invalid simple string: missing CRLF")
)
