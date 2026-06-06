package core

import (
	"time"

	"github.com/ArditZubaku/kvx/config"
)

// TODO: Benchmark and see if we need to introduce
// an objPool at some point or even switch to value semantics

var store map[string]*Obj

func init() {
	// we could include this in the main function,
	// but I think this keeps it clearer
	store = make(map[string]*Obj)
}

func NewObj(value any, expirationMs int64, objType uint8, objEnc uint8) *Obj {
	var expiresAt int64 = -1

	if expirationMs > 0 {
		expiresAt = time.Now().UnixMilli() + expirationMs
	}

	return &Obj{
		Value:        value,
		ExpiresAt:    expiresAt,
		TypeEncoding: objType | objEnc, // 4 first bits objType, last 4 objEnc
	}
}

func Put(key string, value *Obj) {
	if len(store) >= config.KeysLimit {
		evict()
	}
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

	// if the key is expired - delete it and return nil
	// this is the lazy way of deleting expired keys
	if v != nil {
		// we need the -1 check to avoid deleting keys that do not have an expiration
		if v.ExpiresAt != -1 && v.ExpiresAt <= time.Now().UnixMilli() {
			delete(store, key)
			return nil
		}
	}

	return v
}

func Del(key string) bool {
	if _, ok := store[key]; ok {
		delete(store, key)
		KeySpaceStat[0]["keys"]--
		return true
	}
	return false
}
