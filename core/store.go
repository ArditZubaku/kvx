package core

import (
	"time"

	"github.com/ArditZubaku/kvx/config"
)

// TODO: Benchmark and see if we need to introduce
// an objPool at some point or even switch to value semantics

var store map[string]*Obj

type Obj struct {
	Value     any
	ExpiresAt int64
}

func init() {
	// we could include this in the main function,
	// but I think this keeps it clearer
	store = make(map[string]*Obj)
}

func NewObj(value any, expirationMs int64) *Obj {
	var expiresAt int64 = -1

	if expirationMs > 0 {
		expiresAt = time.Now().UnixMilli() + expirationMs
	}

	return &Obj{
		Value:     value,
		ExpiresAt: expiresAt,
	}
}

func Put(key string, value *Obj) {
	if len(store) >= config.KeysLimit {
		evict()
	}
	store[key] = value
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
		return true
	}
	return false
}
