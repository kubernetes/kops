/*
Copyright 2016 The Kubernetes Authors.

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

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

// A test to make sure that fmt.Sprintf on a big.Int is the same as Text
func TestBigInt_Format(t *testing.T) {
	rnd := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	var limit big.Int
	limit.Lsh(big.NewInt(1), 100)
	for i := 1; i < 100; i++ {
		var r big.Int
		r.Rand(rnd, &limit)
		s1 := r.String()
		s2 := r.Text(10)

		fmt.Printf("%s\n", s1)
		if s1 != s2 {
			t.Fatalf("%s not the same as %s", s1, s2)
		}
	}
}
