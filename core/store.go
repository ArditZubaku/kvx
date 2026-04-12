package core

import "time"

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
	store[key] = value
}

func Get(key string) *Obj {
	return store[key]
}

func Del(key string) bool {
	if _, ok := store[key]; ok {
		delete(store, key)
		return true
	}
	return false
}
