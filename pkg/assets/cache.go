/*
Copyright 2018 The Kubernetes Authors.

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

package assets

import (
	"sync"

	"k8s.io/kops/util/pkg/hashing"
)

type assetCache struct {
	mutex  sync.Mutex
	hashes map[string]*hashing.Hash
}

var cache assetCache

func init() {
	cache.hashes = make(map[string]*hashing.Hash)
}

func getCachedHash(u string) *hashing.Hash {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	return cache.hashes[u]
}

func setCachedHash(u string, h *hashing.Hash) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.hashes[u] = h
}
