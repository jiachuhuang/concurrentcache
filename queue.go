package concurrentcache

type Queue struct {
	head *QNode
	tail *QNode
}

type QNode struct {
	prev *QNode
	next *QNode
	v interface{}
}

