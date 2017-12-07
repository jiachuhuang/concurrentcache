package concurrentcache

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

const pickNum = 3

// ConcurrentCache ...
type ConcurrentCache struct {
	segment []*Segment
	sCount  uint32
}

// Segment ..
type Segment struct {
	sync.RWMutex
	data    map[string]*Node
	lfuPool map[string]*Node
	dCount  uint32
	dLen    uint32
	pool    *sync.Pool
	hits    uint64
	miss    uint64
	now     time.Time
}

// Node is cache node
type Node struct {
	V          interface{}
	visit      uint32
	lifeExp    time.Duration
	createTime time.Time
}

// NewConcurrentCache init ConcurrentCache
// sCount is segment num
// dCount is data num of one segment
func NewConcurrentCache(sCount, dCount uint32) (*ConcurrentCache, error) {
	if sCount < 32 || sCount > 256 {
		return nil, errors.New("sCount[Segment num] must be [32,256]")
	}
	if dCount < 1024 || dCount > 65536 {
		return nil, errors.New("dCount[Segment data num] must be [1024,65536]")
	}
	cc := &ConcurrentCache{segment: make([]*Segment, sCount), sCount: sCount}
	for k := range cc.segment {
		cs := newSegment(dCount)
		cc.segment[k] = cs
	}
	return cc, nil
}

func newSegment(dCount uint32) *Segment {
	pool := &sync.Pool{
		New: func() interface{} {
			return &Node{}
		},
	}
	return &Segment{lfuPool: make(map[string]*Node, pickNum), pool: pool, dCount: dCount, data: make(map[string]*Node), dLen: 0}
}

// Set value to the cache
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

// Get value from cache
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

// Delete value from cache
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

// Expire set the value with a expire time
func (cc *ConcurrentCache) Expire(key string, expire time.Duration) (bool, error) {
	if key == "" {
		return false, errors.New("key can not be empty")
	}
	h := MurmurHash2(key)
	h = h % cc.sCount
	cs := cc.segment[h]

	cs.Lock()
	defer cs.Unlock()

	cs.now = time.Now()
	cn, exists := cs.data[key]
	if !exists {
		return true, nil
	}
	if cn.expire(cs.now) || cn.createTime.Add(expire).Before(cs.now) {
		return true, nil
	}
	cn.lifeExp = expire
	atomic.AddUint32(&cn.visit, 1)
	return true, nil
}

// Add k-v to the cache , if the key not exists
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

func (cs *Segment) set(key string, value interface{}, expire time.Duration, nx bool) (bool, error) {
	cs.Lock()
	defer cs.Unlock()

	cs.now = time.Now()
	var cn *Node
	cn, exists := cs.data[key]
	if nx && exists && !cn.expire(cs.now) {
		return false, nil
	}
	if !exists {
		if cs.dLen >= cs.dCount {
			pk := cs.pick()
			cn = cs.data[pk]
		} else {
			cn = cs.newNode()
			cs.dLen++
		}
	}
	cn.reset()
	cn.V = value
	cn.lifeExp = expire
	cs.data[key] = cn
	return true, nil
}

func (cs *Segment) pick() string {
again:
	pl := len(cs.lfuPool)
	for k, v := range cs.data {
		if pl >= pickNum {
			break
		}
		_, exists := cs.lfuPool[k]
		if !exists {
			cs.lfuPool[k] = v
			pl++
		}
	}
	var pk string
	var pkCn *Node
	for k, v := range cs.lfuPool {
		_, exists := cs.data[k]
		if !exists {
			delete(cs.lfuPool, k)
			continue
		}
		if pkCn == nil {
			if v.expire(cs.now) {
				delete(cs.lfuPool, k)
				return k
			}
			pkCn = v
			pk = k
			continue
		}
		if v.expire(cs.now) {
			delete(cs.lfuPool, k)
			pkCn = v
			pk = k
			return k
		}

		if v.visit < pkCn.visit {
			pkCn = v
			pk = k
		}
	}
	if pkCn == nil {
		goto again
	} else {
		delete(cs.lfuPool, pk)
	}
	return pk
}

func (cs *Segment) get(key string) (interface{}, error) {
	cs.RLock()

	cs.now = time.Now()
	cn, exists := cs.data[key]
	if !exists {
		cs.RUnlock()
		atomic.AddUint64(&cs.miss, 1)
		return nil, nil
	}
	if cn.expire(cs.now) {
		cs.RUnlock()
		atomic.AddUint64(&cs.miss, 1)
		return nil, nil
	}
	cs.RUnlock()

	atomic.AddUint32(&cn.visit, 1)
	atomic.AddUint64(&cs.hits, 1)
	return cn.V, nil
}

func (cs *Segment) delete(key string) (bool, error) {
	cs.Lock()
	defer cs.Unlock()

	cn, exists := cs.data[key]
	if !exists {
		return true, nil
	}
	delete(cs.data, key)
	if _, ok := cs.lfuPool[key]; ok {
		delete(cs.lfuPool, key)
	}
	cs.dLen--
	cs.recycle(cn)
	return true, nil
}

func (cs *Segment) newNode() *Node {
	cn := cs.pool.Get().(*Node)
	return cn
}

func (cs *Segment) recycle(cn *Node) {
	if cn != nil {
		cs.pool.Put(cn)
	}
}

func (cn *Node) reset() {
	if cn != nil {
		cn.V = nil
		cn.createTime = time.Now()
		cn.lifeExp = 0
		cn.visit = 0
	}
}

func (cn *Node) expire(now time.Time) bool {
	if cn.lifeExp == 0 {
		return false
	}
	return cn.createTime.Add(cn.lifeExp).Before(now)
}
