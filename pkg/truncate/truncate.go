/*
Copyright 2021 The Kubernetes Authors.

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

package truncate

import (
	"encoding/base32"
	"hash/fnv"
	"strings"

	"k8s.io/klog/v2"
)

// TruncateStringOptions contains parameters for how we truncate strings
type TruncateStringOptions struct {
	// AlwaysAddHash will always cause the hash to be appended.
	// Useful to stop users assuming that the name will never be truncated.
	AlwaysAddHash bool

	// MaxLength controls the maximum length of the string.
	MaxLength int

	// HashLength controls the length of the hash to be appended.
	HashLength int
}

// TruncateString will attempt to truncate a string to a max, adding a prefix to avoid collisions.
// Will never return a string longer than maxLength chars
func TruncateString(s string, opt TruncateStringOptions) string {
	if opt.MaxLength == 0 {
		klog.Fatalf("MaxLength must be set")
	}

	if !opt.AlwaysAddHash && len(s) <= opt.MaxLength {
		return s
	}

	if opt.HashLength == 0 {
		opt.HashLength = 6
	}

	hashString := HashString(s, opt.HashLength)

	maxBaseLength := opt.MaxLength - len(hashString) - 1
	if len(s) > maxBaseLength {
		s = s[:maxBaseLength]
	}
	s = s + "-" + hashString

	return s
}

// HashString will attempt to hash the string.
// Will never return a string longer than length
func HashString(s string, length int) string {
	if length == 0 {
		klog.Fatalf("hash length must be a positive number")
	}

	// We always compute the hash and add it, lest we trick users into assuming that we never do this
	h := fnv.New32a()
	if _, err := h.Write([]byte(s)); err != nil {
		klog.Fatalf("error hashing values: %v", err)
	}
	hashString := base32.HexEncoding.EncodeToString(h.Sum(nil))
	hashString = strings.ToLower(hashString)
	if len(hashString) > length {
		hashString = hashString[:length]
	}

	return hashString
}
