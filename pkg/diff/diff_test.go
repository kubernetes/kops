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

package diff

import (
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func Test_Diff_1(t *testing.T) {
	l := `A
B
C
D
E
F`
	r := `A
D
E
F`
	expectedDiff := `  A
- B
- C
  D
  E
...
`

	{
		dl := buildDiffLines(l, r)
		actual := ""
		for i := range dl {
			switch dl[i].Type {
			case diffmatchpatch.DiffDelete:
				actual += "-"
			case diffmatchpatch.DiffInsert:
				actual += "+"
			case diffmatchpatch.DiffEqual:
				actual += "="
			default:
				t.Errorf("Unexpected diff type: %v", dl[i].Type)
			}
		}
		expected := "=--==="
		if actual != expected {
			t.Fatalf("unexpected diff.  expected=%v, actual=%v", expected, actual)
		}
	}

	actual := FormatDiff(l, r)
	if actual != expectedDiff {
		t.Fatalf("unexpected diff.  expected=%v, actual=%v", expectedDiff, actual)
	}
}

func Test_Diff_2(t *testing.T) {
	l := `A
B
C
D
E
F`
	r := `A
B
C
D
E2
F`
	expectedDiff := `...
  C
  D
+ E2
- E
  F
`
	actual := FormatDiff(l, r)
	if actual != expectedDiff {
		t.Fatalf("unexpected diff.  expected=%v, actual=%v", expectedDiff, actual)
	}
}

func Test_Diff_3(t *testing.T) {
	l := `A
B
C
D
E
F`
	r := `A
B
C
D
E
F2`
	expectedDiff := `...
  D
  E
- F
+ F2
`
	actual := FormatDiff(l, r)
	if actual != expectedDiff {
		t.Fatalf("unexpected diff.  expected=%v, actual=%v", expectedDiff, actual)
	}
}

func Test_Diff_ChangedLine(t *testing.T) {
	l := `ABC123
Line2
Line3`
	r := `ABCDEF
Line2
Line3`
	expectedDiff := `+ ABCDEF
- ABC123
  Line2
  Line3
`
	actual := FormatDiff(l, r)
	if actual != expectedDiff {
		t.Fatalf("unexpected diff.  expected=%v, actual=%v", expectedDiff, actual)
	}
}

func Test_Diff_4(t *testing.T) {
	l := `A
B
C
D
E
F
`
	r := `A
B
C
D
E
F`
	expectedDiff := `...
  D
  E
- F
+ F
`
	{
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(l, r, false)

		// We do need to cleanup, as otherwise we get some spurious changes on complex diffs
		dmp.DiffCleanupSemantic(diffs)

	}

	actual := FormatDiff(l, r)
	if actual != expectedDiff {
		t.Fatalf("unexpected diff.  expected=%s, actual=%s", expectedDiff, actual)
	}
}
