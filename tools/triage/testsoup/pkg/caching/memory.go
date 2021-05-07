package caching

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang/protobuf/proto"
	"k8s.io/klog/v2"
)

var Default Cache

type Cache interface {
	Get(key ...string) CacheEntry
}
type CacheEntry interface {
	// Gets the cache value and deserializes it into msg, or evaluates f.
	// f will normally populate msg.
	GetOrEval(msg proto.Message, f func() error) error
}

func init() {
	cacheDir := os.Getenv("CACHE_DIR")
	if cacheDir == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			klog.Fatalf("os.UserHomeDir failed: %v", err)
		}
		cacheDir = filepath.Join(homedir, ".cache", "testsoup", "memoized")
	}

	c, err := NewDiskCache(cacheDir)
	if err != nil {
		klog.Fatalf("NewDiskCache failed: %v", err)
	}
	Default = c
}

type memoryCache struct {
	mutex   sync.Mutex
	entries map[string]*memoryCacheEntry
}

type memoryCacheEntry struct {
	key   []string
	value []byte
}

func computeHashHex(key []string) string {
	hasher := sha256.New()
	for _, s := range key {
		hasher.Write([]byte(s))
		hasher.Write([]byte{0})
	}
	hash := hasher.Sum(nil)

	hashHex := hex.EncodeToString(hash)
	return hashHex
}

func (c *memoryCache) Get(key ...string) CacheEntry {
	hashHex := computeHashHex(key)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry := c.entries[hashHex]
	if entry != nil {
		if !keyEquals(entry.key, key) {
			klog.Fatalf("hash collision on key %v vs %v", entry.key, key)
		}
		return entry
	}
	entry = &memoryCacheEntry{
		key: key,
	}
	c.entries[hashHex] = entry
	return entry
}

func (e *memoryCacheEntry) Read(msg proto.Message) bool {
	if e.value == nil {
		return false
	}
	if err := proto.Unmarshal(e.value, msg); err != nil {
		klog.Warningf("failed to unmarshal key %v: %v", e.key, err)
		return false
	}
	return true
}

func (e *memoryCacheEntry) GetOrEval(msg proto.Message, f func() error) error {
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

func (e *memoryCacheEntry) Set(msg proto.Message) {
	b, err := proto.Marshal(msg)
	if err != nil {
		klog.Warningf("failed to marshal key %v: %v", e.key, err)
		return
	}
	e.value = b
}

func keyEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
