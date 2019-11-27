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

package utils

import (
	"crypto/sha1"
	"encoding/hex"
)

// HashString Takes String and returns a sha1 hash represented as a string
func HashString(s string) (string, error) {
	h := sha1.New()
	_, err := h.Write([]byte(s))
	if err != nil {
		return "", err
	}
	sha := h.Sum(nil)                 // "sha" is uint8 type, encoded in base16
	shaStr := hex.EncodeToString(sha) // String representation
	return shaStr, nil
}
