package core

// https://redis.io/docs/latest/operate/rs/databases/memory-performance/eviction-policy/
// I will be writing a much simpler eviction algorithm and lru lfu and such
// Simple first - whenever our cache is full, I'll evict the first key that I find

// TODO: Make the eviction strategy config drive
// TODO: Support multiple eviction strategies
func evict() {
	evictFirst()
}

// evictFirst - Evicts the first key it finds while iterating the map
// TODO: Improve via thorough sampling
func evictFirst() {
	for k := range store {
		delete(store, k)
		return
	}
}
