# concurrentcache
concurrentcache是一款golang的内存缓存库，多Segment设计，支持不同Segment间并发写入，提高读写性能。

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
