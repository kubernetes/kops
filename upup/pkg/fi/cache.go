package fi

import "sync"

type Cache interface {
}

type Item struct {
	Object interface{}
}

type cache struct {
	items map[string]Item
	m     sync.RWMutex
}

func New() Cache {
	return &cache{
		items: make(map[string]Item),
	}
}

func (c *cache) Set(k string, v interface{}) {
	c.m.Lock()
	defer c.m.Unlock()

	c.items[k] = Item{
		Object: v,
	}
}

func (c *cache) Get(k string) (interface{}, bool) {
	c.m.RLock()
	defer c.m.RUnlock()

	item, found := c.items[k]
	if !found {
		return nil, false
	}

	return item.Object, true
}
