/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
