package concurrentcache

import (
	"math/rand"
	"strconv"
	"testing"

	"time"
)

func TestNewConcurrentCache(t *testing.T) {
	_, err := NewConcurrentCache(128, 1024)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("success")
	}
}

func TestConcurrentCache_Add(t *testing.T) {
	cc, err := NewConcurrentCache(128, 1024)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("success")
	}

	for try := 10; try > 0; try-- {
		ok, err := cc.Add("abc", 564, 0)
		if err != nil {
			t.Error(err)
		} else if !ok && err == nil {
			t.Log("yeah")
		} else {
			t.Log("oh yeah")
		}
	}
	t.Log("success")
}

func TestConcurrentCache_Set(t *testing.T) {
	cc, err := NewConcurrentCache(128, 1024)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("success")
	}

	for try := 10; try > 0; try-- {
		ok, err := cc.Set("abc", 564, 0)
		if err != nil {
			t.Error(err)
		} else if !ok && err == nil {
			t.Log("yeah")
		} else {
			t.Log("oh yeah")
		}
	}
	t.Log("success")
}

func TestConcurrentCache_Get(t *testing.T) {
	cc, err := NewConcurrentCache(128, 1024)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("success")
	}

	ok, err := cc.Set("abc", 564, 0)
	if err != nil {
		t.Error(err)
	} else if !ok && err == nil {
		t.Log("yeah")
	} else {
		t.Log("oh yeah")
	}

	for try := 5; try > 0; try-- {
		v, err := cc.Get("abc")
		if err != nil {
			t.Error(err)
		} else {
			t.Log(v.(int))
		}
	}
}

func TestConcurrentCache_Delete(t *testing.T) {
	cc, err := NewConcurrentCache(128, 1024)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("success")
	}

	ok, err := cc.Set("abc", 564, 0)
	if err != nil {
		t.Error(err)
	} else if !ok && err == nil {
		t.Log("yeah")
	} else {
		t.Log("oh yeah")
	}

	for try := 5; try > 0; try-- {
		v, err := cc.Get("abc")
		if err != nil {
			t.Error(err)
		} else if v != nil {
			t.Log(v.(int))
		}
	}
	cc.Delete("abc")
	cc.Delete("efg")
	for try := 5; try > 0; try-- {
		v, err := cc.Get("abc")
		if err != nil {
			t.Error(err)
		} else if v != nil {
			t.Log(v.(int))
		}
	}
}

func BenchmarkConcurrentCache_Set(b *testing.B) {
	cc, _ := NewConcurrentCache(256, 10240)

	b.RunParallel(func(pb *testing.PB) {
		var s string
		for pb.Next() {
			i := rand.Int()
			s = strconv.Itoa(i)
			cc.Set(s, s, 5*time.Second)
		}
	})
}

func BenchmarkConcurrentCache_Get(b *testing.B) {
	cc, _ := NewConcurrentCache(256, 10240)
	var s string
	for i := 0; i < 100000; i++ {
		i := rand.Int()
		s = strconv.Itoa(i)
		cc.Set(s, s, 5*time.Second)
	}

	b.RunParallel(func(pb *testing.PB) {
		var s string
		for pb.Next() {
			i := rand.Int()
			s = strconv.Itoa(i)
			cc.Get(s)
		}
	})
}

func TestMurmurHash2(t *testing.T) {
	t.Log(MurmurHash2("a"))
	t.Log(MurmurHash2("ab"))
	t.Log(MurmurHash2("abc"))
	t.Log(MurmurHash2("abcd"))
	t.Log(MurmurHash2("abcD"))
}
