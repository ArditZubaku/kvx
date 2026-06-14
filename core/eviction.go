package core

import (
	"time"

	"github.com/ArditZubaku/kvx/config"
)

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
	case "all-keys-lru":
		evictAllKeysLRU()
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

func evictAllKeysLRU() {
	evictCount := int(config.EvictionRatio * float64(config.KeysLimit))

	if evictCount >= len(evictionPool.keys) {
		populateEvictionPool()
	}

	for i := 0; i < evictCount && len(evictionPool.pool) > 0; i++ {
		item := evictionPool.Pop()
		if item == nil {
			return
		}

		Del(item.key)
	}
}

// --- The approximated LRU algorithm ---
const Max24BitValue = 0x00FFFFFF

// getCurrentClock - returns the 24 bits representing the time in that point
func getCurrentClock() uint32 {
	return uint32(time.Now().Unix()) & Max24BitValue
}

// getIdleTime - the amount of time an obj has been sitting around since the last time it was accessed.
// Accounts for a 24-bit clock rollover.
func getIdleTime(lastAccessedAt uint32) uint32 {
	c := getCurrentClock()
	if c >= lastAccessedAt {
		return c - lastAccessedAt
	}
	return (Max24BitValue - lastAccessedAt) + c
}

func populateEvictionPool() {
	// TODO: Make this configurable
	sampleSize := 5

	for key := range store {
		evictionPool.Push(key, store[key].LastAccessedAt)
		sampleSize--
		if sampleSize == 0 {
			break
		}
	}
}
