package concurrentcache

import (
	"sync"
	"time"
	"errors"
	"sync/atomic"
)

type ConcurrentCache struct {
	segment []*ConcurrentCacheSegment
	sCount uint32
}

type ConcurrentCacheSegment struct {
	sync.RWMutex
	data map[string]*ConcurrentCacheNode
	lvPool map[string]*ConcurrentCacheNode
	dCount uint32
	dLen uint32
	pool *sync.Pool
	hits uint64
	miss uint64
}

type ConcurrentCacheNode struct {
	V interface{}
	visit uint32
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
	pool := &sync.Pool{
		New: func() interface{} {
			return &ConcurrentCacheNode{}
		},
	}
	return &ConcurrentCacheSegment{lvPool:make(map[string]*ConcurrentCacheNode, 5), pool:pool, dCount:dCount, data: make(map[string]*ConcurrentCacheNode), dLen:0}
}

func (cc *ConcurrentCache) Set(key string, value interface{}, expire time.Duration) (bool, error) {
	if key == "" || value == nil {
		return false, errors.New("key or value can not be empty")
	}
	h := MurmurHash2(key)
	h = h % cc.sCount
	cs := cc.segment[h]
	_, err := cs.set(key, value, expire, false)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (cc *ConcurrentCache) Get(key string) (interface{}, error) {
	if key == "" {
		return false, errors.New("key can not be empty")
	}
	h := MurmurHash2(key)
	h = h % cc.sCount
	cs := cc.segment[h]
	value, err := cs.get(key)
	return value, err
}

func (cc *ConcurrentCache) Delete(key string) (bool, error) {
	if key == "" {
		return false, errors.New("key can not be empty")
	}
	h := MurmurHash2(key)
	h = h % cc.sCount
	cs := cc.segment[h]
	_, err := cs.delete(key)
	return true, err
}

func (cc *ConcurrentCache) Expire(key string, expire time.Duration) (bool, error) {
	if key == "" {
		return false, errors.New("key can not be empty")
	}
	h := MurmurHash2(key)
	h = h % cc.sCount
	cs := cc.segment[h]

	cs.Lock()
	defer cs.Unlock()

	cn, exists := cs.data[key]
	if !exists {
		return true, nil
	} else {
		if cn.expire() || cn.createTime.Add(expire).Before(time.Now()) {
			return true, nil
		} else {
			cn.lifeExp = expire
			atomic.AddUint32(&cn.visit, 1)
		}
	}
	return true, nil
}

func (cc *ConcurrentCache) Add(key string, value interface{}, expire time.Duration) (bool, error) {
	if key == "" || value == nil {
		return false, errors.New("key or value can not be empty")
	}
	h := MurmurHash2(key)
	h = h % cc.sCount
	cs := cc.segment[h]
	ok, err := cs.set(key, value, expire, true)
	if !ok && err == nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (cs *ConcurrentCacheSegment) set(key string, value interface{}, expire time.Duration, nx bool) (bool, error) {
	cs.Lock()
	defer cs.Unlock()

	var cn *ConcurrentCacheNode
	cn, exists := cs.data[key]
	if nx && exists && !cn.expire() {
		return false, nil
	}
	if !exists {
		if cs.dLen >= cs.dCount {
			pk := cs.pick()
			cn = cs.data[pk]
		} else {
			cn = cs.newConcurrentCacheNode()
			cs.dLen++
		}
	}
	cn.reset()
	cn.V = value
	cn.lifeExp = expire
	cs.data[key] = cn
	return true, nil
}

func (cs *ConcurrentCacheSegment) pick() string {
	again:
	pl := len(cs.lvPool)
	for k, v := range cs.data {
		if pl >= 5 {
			break
		}
		_, exists := cs.lvPool[k]
		if !exists {
			cs.lvPool[k] = v
			pl++
		}
	}
	var pk string
	var pk_cn *ConcurrentCacheNode
	for k, v := range cs.lvPool {
		_, exists := cs.data[k]
		if !exists {
			delete(cs.lvPool, k)
			continue
		}
		if pk_cn == nil {
			if v.expire() {
				delete(cs.lvPool, k)
				return k
			}
			pk_cn = v
			pk = k
			continue
		}
		if v.expire() {
			delete(cs.lvPool, k)
			pk_cn = v
			pk = k
			return k
		} else {
			if v.visit < pk_cn.visit {
				pk_cn = v
				pk = k
			}
		}
	}
	if pk_cn == nil {
		goto again
	} else {
		delete(cs.lvPool, pk)
	}
	return pk
}

func (cs *ConcurrentCacheSegment) get(key string) (interface{}, error) {
	cs.RLock()
	defer cs.RUnlock()

	cn, exists := cs.data[key]
	if !exists {
		return nil, nil
	}
	if cn.expire() {
		return nil, nil
	} else {
		atomic.AddUint32(&cn.visit, 1)
		return cn.V, nil
	}
}

func (cs *ConcurrentCacheSegment) delete(key string) (bool, error) {
	cs.Lock()
	defer cs.Unlock()

	cn, exists := cs.data[key]
	if !exists {
		return true, nil
	}
	delete(cs.data, key)
	if _, ok := cs.lvPool[key]; ok {
		delete(cs.lvPool, key)
	}
	cs.dLen--
	cs.pool.Put(cn)
	return true, nil
}

func (cs *ConcurrentCacheSegment) newConcurrentCacheNode() *ConcurrentCacheNode {
	cn := cs.pool.Get().(*ConcurrentCacheNode)
	return cn
}

func (cs *ConcurrentCacheSegment) recycle(cn *ConcurrentCacheNode) {
	if cn != nil {
		cs.pool.Put(cn)
	}
}

func (cn *ConcurrentCacheNode) reset() {
	if cn != nil {
		cn.V = nil
		cn.createTime = time.Now()
		cn.lifeExp = 0
		cn.visit = 0
	}
}

func (cn *ConcurrentCacheNode) expire() bool {
	if cn.lifeExp == 0 {
		return false
	}
	return cn.createTime.Add(cn.lifeExp).Before(time.Now())
}