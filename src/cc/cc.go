package cc

import (
	"sync"
)

type Lru struct {
	store      map[string]lruItem
	lock       *sync.Mutex
	head, tail *lruNode
	cap, len   int
}

type lruItem struct {
	value interface{}
	node  *lruNode
}

type lruNode struct {
	key        string
	prev, next *lruNode
}

func NewLru(capacity int) *Lru {
	if capacity < 2 {
		capacity = 2 // helpful invariant
	}
	return &Lru{
		store: make(map[string]lruItem, capacity),
		lock:  &sync.Mutex{},
		head:  nil,
		tail:  nil,
		cap:   capacity,
		len:   0,
	}
}

func (l *Lru) Clear() {
	l.lock.Lock()
	defer l.lock.Unlock()
	for k, _ := range l.store {
		delete(l.store, k)
	}
	l.head, l.tail, l.len = nil, nil, 0
}

func (l *Lru) Set(k string, v interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.remove(k) // note "remove", not "Remove"
	l.set(k, v)
}

func (l *Lru) set(k string, v interface{}) {

	var n *lruNode

	if l.len < l.cap {

		n = &lruNode{k, nil, l.head}
		if l.head != nil {
			l.head.prev = n
		}
		l.head = n
		if l.tail == nil {
			l.tail = n
		}
		l.len++

	} else {

		// pop tail off
		r := l.tail
		p := r.prev
		p.next = nil
		r.prev = nil
		l.tail = p
		delete(l.store, r.key)

		// push head in
		n = &lruNode{k, nil, l.head}
		l.head.prev = n
		l.head = n

	}

	l.store[k] = lruItem{v, n}
}

func (l *Lru) Get(k string) (interface{}, bool) {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.get(k)
}

func (l *Lru) get(k string) (interface{}, bool) {

	i, b := l.store[k]

	if !b {
		return nil, false
	}

	p := i.node.prev
	n := i.node.next

	if p == nil { // already at head
		return i.value, b
	}

	if n == nil { // at tail
		p.next = nil
		l.tail = p
	} else { // somewhere in between
		p.next = n
		n.prev = p
	}

	i.node.prev = nil
	l.head.prev = i.node
	i.node.next = l.head
	l.head = i.node

	return i.value, b
}

func (l *Lru) Remove(k string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.remove(k)
}

func (l *Lru) remove(k string) {

	i, b := l.store[k]

	if !b {
		return
	}

	delete(l.store, k)

	p := i.node.prev
	n := i.node.next

	if p == nil { // at head
		if n != nil {
			n.prev = nil
		}
		l.head = n
		l.len--
		return
	}

	if n == nil { // at tail
		p.next = nil
		l.tail = p
		l.len--
		return
	}

	// somewhere in between
	p.next = n
	n.prev = p
	l.len--

}
