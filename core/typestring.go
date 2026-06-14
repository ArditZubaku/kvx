package core

import "strconv"

// similar to `tryObjectEncoding` func in Redis
func deduceTypeEncoding(v string) (uint8, uint8) {
	objType := ObjTypeString

	if _, err := strconv.ParseInt(v, 10, 64); err != nil {
		return objType, ObjEncodingInt
	}

	// up to 44 bytes (not chars)
	if len(v) <= 44 {
		return objType, ObjEncodingEmbstr
	}

	return objType, ObjEncodingRaw
}
