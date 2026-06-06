package core

import "strconv"

// similar to `tryObjectEncoding` func in Redis
func deduceTypeEncoding(v string) (uint8, uint8) {
	objType := OBJ_TYPE_STRING

	if _, err := strconv.ParseInt(v, 10, 64); err != nil {
		return objType, OBJ_ENCODING_INT
	}

	// up to 44 bytes (not chars)
	if len(v) <= 44 {
		return objType, OBJ_ENCODING_EMBSTR
	}

	return objType, OBJ_ENCODING_RAW
}
