// core contains everything related to the RESP protocol - Redis serialization protocol
package core

import (
	"bytes"
)

func Decode(data []byte) (any, error) {
	if len(data) == 0 {
		return nil, ErrNoData
	}

	value, _, err := DecodeOne(data)
	return value, err
}

func DecodeOne(data []byte) (any, int, error) {
	if len(data) == 0 {
		return nil, 0, ErrNoData
	}

	switch data[0] {
	case '+':
		return readSimpleString(data)
	case '-':
		return readError(data)
	case ':':
		return readInt64(data)
	case '$':
		return readBulkString(data)
	case '*':
		return readArray(data)
	}

	return nil, 0, nil
}

// readSimpleString - https://redis.io/docs/latest/develop/reference/protocol-spec/#simple-strings
// reads a RESP encoded simple string from data and
// returns the string, the cursor (up to where it read) and the error
func readSimpleString(data []byte) (string, int, error) {
	// Expect: +<string>\r\n
	if len(data) < 3 {
		return "", 0, ErrInvalidSimpleString
	}

	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return "", 0, ErrMissingCRLF
	}

	// TODO: Rethink whether I should return string or []byte
	return string(data[1:idx]), idx + 2, nil
}

// readError - https://redis.io/docs/latest/develop/reference/protocol-spec/#simple-errors
// reads a RESP encodede error from data and
// returns the error string, the cursor (up to where it read) and the error
func readError(data []byte) (string, int, error) {
	return readSimpleString(data)
}

// readInt64 - https://redis.io/docs/latest/develop/reference/protocol-spec/#integers
// reads a RESP encodede error from data and
// returns the integer value, the cursor (up to where it read) and the error
func readInt64(data []byte) (int64, int, error) {
	// Expect: :<string>\r\n
	if len(data) < 3 {
		return 0, 0, ErrInvalidInt64
	}

	pos := 1
	sign := int64(1)

	if pos < len(data) && data[pos] == '-' {
		sign = -1
		pos++
	}

	if pos >= len(data) {
		return 0, 0, ErrInvalidInt64
	}

	var value int64

	for ; data[pos] != '\r'; pos++ {
		c := data[pos]
		if c < '0' || c > '9' {
			return 0, 0, ErrInvalidInt64
		}
		value = value*10 + int64(c-'0')
	}

	// validate CRLF
	if pos+1 >= len(data) || data[pos] != '\r' || data[pos+1] != '\n' {
		return 0, 0, ErrMissingCRLF
	}

	return sign * value, pos + 2, nil
}

// readBulkString - https://redis.io/docs/latest/develop/reference/protocol-spec/#bulk-strings
func readBulkString(data []byte) (any, int, error) {
	panic("unimplemented")
}

// readArray - https://redis.io/docs/latest/develop/reference/protocol-spec/#arrays
func readArray(data []byte) (any, int, error) {
	panic("unimplemented")
}

// TODO: Implement the encodings for the other types as well (like Doubles, Big numbers, Maps...)
