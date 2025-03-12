//go:build !solution

package lrucache

import "container/list"

type KeyValue struct {
	key   int
	value int
}

type LRUCache struct {
	list  *list.List
	items map[int]*list.Element
	cap   int
}

func (c *LRUCache) Get(key int) (int, bool) {
	if element, ok := c.items[key]; ok {
		c.list.MoveToBack(element)
		return element.Value.(KeyValue).value, true
	}
	return 0, false
}

func (c *LRUCache) Set(key, value int) {
	if element, ok := c.items[key]; ok {
		c.list.MoveToBack(element)
		element.Value = KeyValue{key: key, value: value}
		return
	}

	if c.list.Len() >= c.cap {
		if c.cap > 0 && c.list.Len() > 0 {
			oldest := c.list.Front()
			delete(c.items, oldest.Value.(KeyValue).key)
			c.list.Remove(oldest)
		}
	}

	if c.cap > 0 {
		element := c.list.PushBack(KeyValue{key: key, value: value})
		c.items[key] = element
	}
}

func (c *LRUCache) Clear() {
	c.list.Init()
	c.items = make(map[int]*list.Element)
}

func (c *LRUCache) Range(f func(key, value int) bool) {
	for e := c.list.Front(); e != nil; e = e.Next() {
		kv := e.Value.(KeyValue)
		if !f(kv.key, kv.value) {
			break
		}
	}
}

func New(cap int) Cache {
	return &LRUCache{
		list:  list.New(),
		items: make(map[int]*list.Element),
		cap:   cap,
	}
}
