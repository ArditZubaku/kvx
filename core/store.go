package core

import (
	"github.com/ArditZubaku/kvx/config"
)

// TODO: Benchmark and see if we need to introduce
// an objPool at some point or even switch to value semantics

var store map[string]*Obj
var expires map[*Obj]uint64 // ptr->expirationTime

func init() {
	// we could include this in the main function,
	// but I think this keeps it clearer
	store = make(map[string]*Obj)
	expires = make(map[*Obj]uint64)
}

func NewObj(value any, expirationMs int64, objType uint8, objEnc uint8) *Obj {
	obj := &Obj{
		Value:          value,
		TypeEncoding:   objType | objEnc, // 4 first bits objType, last 4 objEnc
		LastAccessedAt: getCurrentClock(),
	}

	if expirationMs > 0 {
		setExpiry(obj, expirationMs)
	}

	return obj
}

func Put(key string, value *Obj) {
	if len(store) >= config.KeysLimit {
		evict()
	}

	value.LastAccessedAt = getCurrentClock()

	store[key] = value

	// NOTE: might remove it in the future and just use len(store) for keys
	// Everything by default goes to db-0
	if KeySpaceStat[0] == nil {
		KeySpaceStat[0] = make(map[string]int)
	}
	KeySpaceStat[0]["keys"]++
}

func Get(key string) *Obj {
	v := store[key]

	// this is the lazy way of deleting expired keys
	if v != nil {
		if hasExpired(v) {
			Del(key)
			return nil
		}
	}

	v.LastAccessedAt = getCurrentClock()

	return v
}

func Del(key string) bool {
	if obj, ok := store[key]; ok {
		delete(store, key)
		delete(expires, obj)
		KeySpaceStat[0]["keys"]--
		return true
	}
	return false
}
