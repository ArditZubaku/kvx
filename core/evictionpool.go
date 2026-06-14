package core

import "sort"

const evictionPoolMaxSize = 16

var evictionPool *EvictionPool = newEvictionPool(0)

type PoolItem struct {
	key            string
	lastAccessedAt uint32
}

// TODO: When lastAccessedAt of	an object changes
// update the poolItem corresponding to that object
type EvictionPool struct {
	pool []*PoolItem
	keys map[string]*PoolItem
}

func newEvictionPool(size int) *EvictionPool {
	return &EvictionPool{
		pool: make([]*PoolItem, size),
		keys: make(map[string]*PoolItem),
	}
}

// TODO: Make the impl more efficient to not need repeated sorting
func (ep *EvictionPool) Push(key string, lastAccessedAt uint32) {
	_, ok := ep.keys[key]
	if !ok {
		return
	}

	item := &PoolItem{key: key, lastAccessedAt: lastAccessedAt}
	if len(ep.pool) < evictionPoolMaxSize {
		ep.keys[key] = item
		ep.pool = append(ep.pool, item)
		// NOTE: Performance bottleneck
		// Maybe use insertion sort
		sort.Sort(ByIdleTime(ep.pool))
	} else if lastAccessedAt > ep.pool[len(ep.pool)-1].lastAccessedAt {
		ep.pool[0] = nil
		ep.pool = ep.pool[1:]
		ep.keys[key] = item
		ep.pool = append(ep.pool, item)
	}
}

func (ep *EvictionPool) Pop() *PoolItem {
	if len(ep.pool) == 0 {
		return nil
	}

	item := ep.pool[0]
	// Explicitly nil out the reference to prevent memory leaks
	ep.pool[0] = nil
	ep.pool = ep.pool[1:]

	return item
}

type ByIdleTime []*PoolItem

func (this ByIdleTime) Len() int {
	return len(this)
}

func (this ByIdleTime) Swap(a, b int) {
	this[a], this[b] = this[b], this[a]
}

func (this ByIdleTime) Less(a, b int) bool {
	return getIdleTime(this[a].lastAccessedAt) > getIdleTime(this[b].lastAccessedAt)
}
