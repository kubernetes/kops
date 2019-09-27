/*
Copyright 2019 The Kubernetes Authors.

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

package vfs

import (
	"sync"
	"time"
)

// NewCache is a constructor for a Cache
func NewCache() *Cache {
	return &Cache{
		cache: make(map[string]cacheEntry),
	}
}

// Cache is a simple cache for vfs files.
//
// Currently we never expire the cache, so this is only safe for a
// relatively bounded set of files - but it would not be too hard to
// fix this.
type Cache struct {
	mutex sync.Mutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	Added    time.Time
	Contents []byte
}

func (c *Cache) Read(p Path, ttl time.Duration) ([]byte, error) {
	key := p.Path()

	c.mutex.Lock()
	entry, found := c.cache[key]
	c.mutex.Unlock()

	// Treat expired as not-found
	if found {
		expiresAt := entry.Added.Add(ttl)
		if time.Now().After(expiresAt) {
			found = false
		}
	}

	if found {
		return entry.Contents, nil
	}

	b, err := p.ReadFile()
	if err != nil {
		return nil, err
	}

	entry.Contents = b
	entry.Added = time.Now()

	c.mutex.Lock()
	c.cache[key] = entry
	c.mutex.Unlock()

	return b, nil
}
