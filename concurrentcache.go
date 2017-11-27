package concurrentcache

import (
	"sync"
	"time"
)

type ConcurrentCache struct {

}

type ConcurrentCacheSegment struct {
	sync.RWMutex
}

type ConcurrentCacheNode struct {
	v interface{}
	lifeTime time.Duration
	createTime time.Time
}