package concurrentcache

import "testing"

func TestNewQueue(t *testing.T) {
	q := NewQueue()
	if q == nil || q.pool == nil {
		t.Error("NewQueue fail")
	} else {
		t.Log("NewQueue ok")
	}
}

func TestQueue_LPush(t *testing.T) {
	q := NewQueue()
	if q == nil || q.pool == nil {
		t.Error("NewQueue fail")
	} else {
		t.Log("NewQueue ok")
	}
	q.LPush(q.NewQNode("a"))
	q.LPush(q.NewQNode("b"))
	q.LPush(q.NewQNode("c"))
	n := q.RPop()
	q.Recycle(n)
	q.LPush(q.NewQNode("d"))
	q.LPush(q.NewQNode("e"))
	n = q.RPop()
	q.Recycle(n)
	q.LPush(q.NewQNode("f"))

	for !q.Empty() {
		t.Log(q.LPop().V.(string))
	}
}

func TestQueue_RPush(t *testing.T) {
	q := NewQueue()
	if q == nil || q.pool == nil {
		t.Error("NewQueue fail")
	} else {
		t.Log("NewQueue ok")
	}
	for i:=100; i>=0; i--{
		q.RPush(q.NewQNode("e"))
	}
	for i:=20; i>=0; i--{
		q.Recycle(q.RPop())
	}
	for i:=50; i>=0; i--{
		q.RPush(q.NewQNode("e"))
	}
	q.RPush(q.NewQNode("a"))
	q.RPush(q.NewQNode("b"))
	q.RPush(q.NewQNode("c"))

	for !q.Empty() {
		t.Log(q.RPop().V.(string))
	}
}

func TestMurmurHash2(t *testing.T) {
	t.Log(MurmurHash2("a"))
	t.Log(MurmurHash2("ab"))
	t.Log(MurmurHash2("abc"))
	t.Log(MurmurHash2("abcd"))
	t.Log(MurmurHash2("ab c"))
	t.Log(MurmurHash2("abcde"))
	t.Log(MurmurHash2("ABCDE"))
	t.Log(MurmurHash2("adfhaksfeuiknjcshfiuenjfharnkj,cjrk"))
	t.Log(MurmurHash2("adfhaksfeuiknjcshfiuenjfharnkj,cjrp"))
}
