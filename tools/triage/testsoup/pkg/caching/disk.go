package caching

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang/protobuf/proto"
	"k8s.io/klog/v2"
)

func NewDiskCache(cacheDir string) (Cache, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("os.MkdirAll(%q) failed: %w", cacheDir, err)
	}
	c := &diskCache{
		cacheDir: cacheDir,
		entries:  make(map[string]*diskCacheEntry),
	}
	return c, nil
}

type diskCache struct {
	cacheDir string

	mutex   sync.Mutex
	entries map[string]*diskCacheEntry
}

type diskCacheEntry struct {
	key           []string
	cacheFilePath string
}

func (c *diskCache) Get(key ...string) CacheEntry {
	hashHex := computeHashHex(key)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	p := filepath.Join(c.cacheDir, hashHex)

	entry := c.entries[hashHex]
	if entry != nil {
		if !keyEquals(entry.key, key) {
			klog.Fatalf("hash collision on key %v vs %v", entry.key, key)
		}
		return entry
	}
	entry = &diskCacheEntry{
		key:           key,
		cacheFilePath: p,
	}
	c.entries[hashHex] = entry
	return entry
}

func (e *diskCacheEntry) Read(msg proto.Message) bool {
	v, err := os.ReadFile(e.cacheFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			klog.Warningf("failed to read file %v: %v", e.cacheFilePath, err)
		}
		return false
	}
	// TODO: Store wrapped including key and verify key
	if err := proto.Unmarshal(v, msg); err != nil {
		klog.Warningf("failed to unmarshal file %v: %v", e.cacheFilePath, err)
		return false
	}
	return true
}

func (e *diskCacheEntry) GetOrEval(msg proto.Message, f func() error) error {
	if e.Read(msg) {
		return nil
	}
	err := f()
	if err != nil {
		return err
	}
	e.Set(msg)
	return nil
}

func (e *diskCacheEntry) Set(msg proto.Message) {
	b, err := proto.Marshal(msg)
	if err != nil {
		klog.Warningf("failed to marshal key %v: %v", e.key, err)
		return
	}

	if err := os.WriteFile(e.cacheFilePath, b, 0644); err != nil {
		klog.Warningf("failed to write cache file %v: %v", e.cacheFilePath, err)
		return
	}
}
