package core

import (
	"log"
	"time"
)

// DeleteExpiredKeys - deletes all the expired keys, the active way
//
// Sampling approach: https://redis.io/docs/latest/commands/expire/
func DeleteExpiredKeys() {
	for {
		frac := expireSample()
		// if the sample has less than 25% of expired keys
		// we break the loop
		if frac < 0.25 {
			break
		}
	}
	log.Println("Deleted the expired keys. Total current keys", len(store))
}

// expireSample
//
// TODO: Optimize sampling and get rid of unnecessary iteration
func expireSample() float32 {
	limit := 20 // try 20 random keys
	expiredCount := 0

	// in Go iteration over map is randomized - https://go.dev/blog/maps#iteration-order
	for key, obj := range store {
		limit--
		if hasExpired(obj) {
			Del(key)
			expiredCount++
		}

		if limit == 0 {
			break
		}
	}

	return float32(expiredCount) / float32(20.0)
}

func setExpiry(obj *Obj, expDurationMs int64) {
	expires[obj] = uint64(time.Now().UnixMilli()) + uint64(expDurationMs)
}

func getExpiry(obj *Obj) (uint64, bool) {
	exp, ok := expires[obj]
	return exp, ok
}

func hasExpired(obj *Obj) bool {
	exp, ok := expires[obj]

	if !ok {
		return false
	}

	return exp <= uint64(time.Now().UnixMilli())
}
