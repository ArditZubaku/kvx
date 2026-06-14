package core

// Obj TODO: change ExpiresAt to LRU Bits as handles by Redis
type Obj struct {
	TypeEncoding uint8
	Value        any
	/*
	 * Redis allocates 24 bits, but we will use 32 bits bc Golang doesn't bit-fields
	 * but we could merge TypeEncoding and LastAccessedAt into one 32-bit integer
	 */
	LastAccessedAt uint32
}

var ObjTypeString uint8 = 0 << 4

var (
	ObjEncodingRaw    uint8 = 0
	ObjEncodingInt    uint8 = 1
	ObjEncodingEmbstr uint8 = 8
)
