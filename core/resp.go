// core contains everything related to the RESP protocol - Redis serialization protocol
package core

import (
	"bytes"
	"math"
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
	isNegative := false
	if data[pos] == '-' {
		isNegative = true
		pos++
	}

	var value int64
	startPos := pos

	for pos < len(data) && data[pos] >= '0' && data[pos] <= '9' {
		digit := int64(data[pos] - '0')

		// Overflow check for negative accumulation
		// We check if value < (MinInt64 + digit) / 10
		if isNegative {
			if value < (math.MinInt64+digit)/10 {
				return 0, 0, ErrIntegerOverflow
			}
			value = value*10 - digit
		} else {
			// For positive, we check against MaxInt64
			if value > (math.MaxInt64-digit)/10 {
				return 0, 0, ErrIntegerOverflow
			}
			value = value*10 + digit
		}
		pos++
	}

	// check if we actually parsed any digits
	if pos == startPos {
		return 0, 0, ErrInvalidInt64
	}

	// validate CRLF
	if pos+1 >= len(data) || data[pos] != '\r' || data[pos+1] != '\n' {
		return 0, 0, ErrMissingCRLF
	}

	return value, pos + 2, nil
}

// readBulkString - https://redis.io/docs/latest/develop/reference/protocol-spec/#bulk-strings
// reads a RESP encodede error from data and
// returns the string, the cursor (up to where it read) and the error
func readBulkString(data []byte) (string, int, error) {
	// Expect: $<intLength>\r\n<stringAsLongAsLength>\r\n
	if len(data) < 4 || data[0] != '$' {
		return "", 0, ErrInvalidBulkString
	}

	pos := 1

	// parse length
	length, cursor, err := readLength(data[pos:])
	if err != nil {
		return "", 0, err
	}
	pos += cursor

	// bounds check for data + CRLF
	if pos+length+2 > len(data) {
		return "", 0, ErrInvalidBulkString
	}

	// extract string
	value := string(data[pos : pos+length])
	pos += length

	// validate trailing CRLF
	if data[pos] != '\r' || data[pos+1] != '\n' {
		return "", 0, ErrMissingCRLF
	}

	return value, pos + 2, nil
}

func readLength(data []byte) (int, int, error) {
	if len(data) == 0 {
		return 0, 0, ErrInvalidInt64
	}

	pos, length := 0, 0

	for pos < len(data) && data[pos] != '\r' {
		c := data[pos]
		if c < '0' || c > '9' {
			return 0, 0, ErrInvalidInt64
		}
		length = length*10 + int(c-'0')
		pos++
	}

	// validate CRLF
	if pos+1 >= len(data) || data[pos] != '\r' || data[pos+1] != '\n' {
		return 0, 0, ErrMissingCRLF
	}

	return length, pos + 2, nil
}

// readArray - https://redis.io/docs/latest/develop/reference/protocol-spec/#arrays
func readArray(data []byte) (any, int, error) {
	panic("unimplemented")
}

// TODO: Implement the encodings for the other types as well (like Doubles, Big numbers, Maps...)
