package concurrentcache

import "sync"

type Queue struct {
	head *QNode
	tail *QNode
	pool *sync.Pool
}

type QNode struct {
	prev *QNode
	next *QNode
	v interface{}
}

