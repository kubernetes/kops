/*
Copyright 2023 The Kubernetes Authors.

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

package mutexes

import (
	"sync"
)

var InProcess LocalMutexes

// LocalMutexes is a store of named mutexes, used to avoid concurrent local operations.
// For example, GCE project IAM mutation uses a read / conditional-write approach,
// so we try to avoid making two local concurrent calls to the same project.
type LocalMutexes struct {
	mutex   sync.Mutex
	mutexes map[string]*localMutex
}

func (m *LocalMutexes) Get(key string) LocalMutex {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.mutexes == nil {
		m.mutexes = make(map[string]*localMutex)
	}

	l := m.mutexes[key]
	if l == nil {
		l = &localMutex{}
		m.mutexes[key] = l
	}
	return l
}

// LocalMutex is the interface for local mutexes.
type LocalMutex interface {
	Lock()
	Unlock()
}

// localMutex implements LocalMutex.
type localMutex struct {
	m sync.Mutex
}

// Lock implements LocalMutex.
func (m *localMutex) Lock() {
	m.m.Lock()
}

// Unlock implements LocalMutex.
func (m *localMutex) Unlock() {
	m.m.Unlock()
}
