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

package tester

import (
	"testing"

	"github.com/urfave/sflags/gen/gpflag"
)

func TestFlagParsing(t *testing.T) {
	tester := &Tester{}

	fs, err := gpflag.Parse(tester)
	if err != nil {
		t.Fatalf("gpflag.Parse(tester) failed: %v", err)
	}

	args := []string{"--parallel", "25"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("fs.Parse(args) failed: %v", err)
	}

	if tester.Parallel != 25 {
		t.Errorf("unexpected value for Parallel; got %d, want %d", tester.Parallel, 25)
	}
}

func TestHasFlag(t *testing.T) {
	grid := []struct {
		Args     string
		Flag     string
		Expected bool
	}{
		{
			Args:     "--provider aws",
			Flag:     "provider",
			Expected: true,
		},
		{
			Args:     "-provider aws",
			Flag:     "provider",
			Expected: true,
		},
		{
			Args:     "provider aws",
			Flag:     "provider",
			Expected: false,
		},
		{
			Args:     "-provider=aws",
			Flag:     "provider",
			Expected: true,
		},
		{
			Args:     "--provider=aws",
			Flag:     "provider",
			Expected: true,
		},
		{
			Args:     "--foo=bar --provider aws",
			Flag:     "provider",
			Expected: true,
		},
		{
			Args:     "--foo=bar",
			Flag:     "provider",
			Expected: false,
		},
	}

	for _, g := range grid {
		t.Run(g.Args, func(t *testing.T) {
			got := hasFlag(g.Args, g.Flag)
			if got != g.Expected {
				t.Errorf("hasFlags(%q, %q) got %v, want %v", g.Args, g.Flag, got, g.Expected)
			}
		})
	}
}
