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

package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {
	m := map[string]bool{
		"key0": true,
		"key1": true,
		"key2": true,
	}
	assert.Equal(t, 3, len(Keys(m)))
}

func TestSortedKeys(t *testing.T) {
	m := map[string]bool{
		"key2": true,
		"key1": true,
		"key0": true,
	}
	assert.Equal(t, []string{"key0", "key1", "key2"}, SortedKeys(m))
}
