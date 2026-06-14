package core

// TODO: change ExpiresAt to LRU Bits as handles by Redis
type Obj struct {
	TypeEncoding uint8
	Value        any
	/*
	 * Redis allocates 24 bits but we will use 32 bits bc Golang doesn't bit-fields
	 * but we could merge TypeEncoding and LastAccessedAt into one 32 bit integer
	 */
	LastAccessedAt uint32
}

var OBJ_TYPE_STRING uint8 = 0 << 4

var OBJ_ENCODING_RAW uint8 = 0
var OBJ_ENCODING_INT uint8 = 1
var OBJ_ENCODING_EMBSTR uint8 = 8
