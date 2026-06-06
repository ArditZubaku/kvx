package core

import "github.com/ArditZubaku/kvx/config"

// https://redis.io/docs/latest/operate/rs/databases/memory-performance/eviction-policy/
// I will be writing a much simpler eviction algorithm than lru lfu and such
// Simple first - whenever our cache is full, I'll evict the first key that I find

// TODO: Make the eviction strategy config drive
// TODO: Support multiple eviction strategies
func evict() {
	switch config.EvictionStrategy {
	case "simple-first":
		evictFirst()
	case "all-keys-random":
		evictAllKeysRandom()
	}
}

// evictFirst - Evicts the first key it finds while iterating the map
// TODO: Improve via thorough sampling
func evictFirst() {
	for k := range store {
		delete(store, k)
		return
	}
}

// evictAllKeysRandom - randomly removes keys to make space for the new data added.
// The number of keys removed will be sufficient to free up at least 10% space
func evictAllKeysRandom() {
	evictCount := int64(config.EvictionRatio * float64(config.KeysLimit))

	// iteration in Go is random
	for k := range store {
		Del(k)
		evictCount--
		if evictCount <= 0 {
			break
		}
	}
}
