// Copyright 2026 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package limit provides bounded reads from io.Readers.
package limit

import (
	"fmt"
	"io"
)

// ReadAll reads from r until EOF and returns the data. If r produces more
// than max bytes, it returns an error instead of a silently truncated
// slice. Use this in preference to io.ReadAll(io.LimitReader(r, max))
// when callers must distinguish a complete read from a truncated one.
func ReadAll(r io.Reader, max int64) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > max {
		return nil, fmt.Errorf("body exceeds %d byte limit", max)
	}
	return b, nil
}
