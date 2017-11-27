package concurrentcache

import (
	"sync"
	"time"
)

type ConcurrentCache struct {
	segment []*ConcurrentCacheSegment
	sCount uint32
}

type ConcurrentCacheSegment struct {
	sync.RWMutex
	lruQueue *Queue
	data map[string]interface{}
	dCount uint32
	dLen uint32
	pool *sync.Pool
}

type ConcurrentCacheNode struct {
	V interface{}
	lifeExp time.Duration
	createTime time.Time
}