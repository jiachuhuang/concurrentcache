package concurrentcache

import (
	"sync"
	"time"
	"errors"
)

type ConcurrentCache struct {
	sync.Mutex
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
	hits uint64
	miss uint64
}

type ConcurrentCacheNode struct {
	V interface{}
	lifeExp time.Duration
	createTime time.Time
}

func NewConcurrentCache(sCount, dCount uint32) (*ConcurrentCache, error) {
	if sCount < 32 || sCount > 128 {
		return nil, errors.New("sCount[ConcurrentCacheSegment num] must be [32,128]")
	}
	if dCount < 1024 || dCount > 65536 {
		return nil, errors.New("dCount[ConcurrentCacheSegment data num] must be [1024,65536]")
	}
	cc := &ConcurrentCache{segment:make([]*ConcurrentCacheSegment, sCount), sCount:sCount}
	for k, _ := range cc.segment {
		cs := newConcurrentCacheSegment(dCount)
		cc.segment[k] = cs
	}
	return cc, nil
}

func newConcurrentCacheSegment(dCount uint32) *ConcurrentCacheSegment {
	lq := NewQueue()
	pool := &sync.Pool{
		New: func() interface{} {
			return &ConcurrentCacheNode{lifeExp:0, createTime: time.Now()}
		},
	}
	return &ConcurrentCacheSegment{lruQueue:lq, pool:pool, dCount:dCount, data: make(map[string]interface{}), dLen:0}
}