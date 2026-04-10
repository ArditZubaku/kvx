package core

import "errors"

var (
	ErrNoData              = errors.New("no data")
	ErrInvalidSimpleString = errors.New("invalid simple string")
	ErrMissingCRLF         = errors.New("invalid simple string: missing CRLF")
)
