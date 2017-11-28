# concurrentcache
concurrentcache是一款golang的内存缓存库，多Segment设计，支持不同Segment间并发写入，提高读写性能。

# Install
```golang
go get github.com/jiachuhuang/concurrentcache
```
Test：
```golang
cd $GOPATH/src/github.com/jiachuhuang/concurrentcache
go test -v -test.bench=".*" -parallel 1000 -count 3 -benchmem
```

# Example
```golang
package main

import (
    "concurrentcache"
	"time"
)
cc, err := NewConcurrentCache(256, 10240)
if err != nil {
    panic(err)
}
ok, err := cc.Set("foo", "bar", 5 * time.Second)
if err != nil {
    panic(err)
}
v, err := cc.Get("foo")
if v != nil {
    fmt.Println(v.(string))
}
```

# Benchmark
concurrentcache和cache2go进行了并发下的压测对比，对比结果，concurrentcache无论是执行时间还是内存占比，都比cache2go优
## concurrentcache
```golang
BenchmarkConcurrentCache_Set-8           3000000               573 ns/op             190 B/op          2 allocs/op
BenchmarkConcurrentCache_Set-8           2000000               634 ns/op             235 B/op          3 allocs/op
BenchmarkConcurrentCache_Set-8           3000000               535 ns/op             190 B/op          2 allocs/op
BenchmarkConcurrentCache_Get-8           5000000               235 ns/op              36 B/op          1 allocs/op
BenchmarkConcurrentCache_Get-8           5000000               234 ns/op              36 B/op          1 allocs/op
BenchmarkConcurrentCache_Get-8           5000000               234 ns/op              36 B/op          1 allocs/op
```

## cache2go
```golang
BenchmarkCacheTable_Add-8     	 1000000	      1858 ns/op	     480 B/op	      10 allocs/op
BenchmarkCacheTable_Value-8   	 5000000	       300 ns/op	      60 B/op	       2 allocs/op
```