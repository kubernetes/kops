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
	"bytes"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	"k8s.io/klog"
)

func FormatDiff(lString, rString string) string {
	results := buildDiffLines(lString, rString)

	return renderText(results, 2)
}

func renderText(results []lineRecord, context int) string {
	keep := make([]bool, len(results))
	for i := range results {
		if results[i].Type == diffmatchpatch.DiffEqual {
			continue
		}
		for j := i - context; j <= i+context; j++ {
			if j >= 0 && j < len(keep) {
				keep[j] = true
			}
		}
	}

	var b bytes.Buffer
	wroteSkip := false
	for i := range results {
		if !keep[i] {
			if !wroteSkip {
				b.WriteString("...\n")
				wroteSkip = true
			}
			continue
		}

		switch results[i].Type {
		case diffmatchpatch.DiffDelete:
			b.WriteString("- ")
		case diffmatchpatch.DiffInsert:
			b.WriteString("+ ")
		case diffmatchpatch.DiffEqual:
			b.WriteString("  ")
		}
		b.WriteString(results[i].Line)
		b.WriteString("\n")
		wroteSkip = false
	}

	return b.String()
}

type lineRecord struct {
	Type diffmatchpatch.Operation
	Line string
}

func buildDiffLines(lString, rString string) []lineRecord {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(lString, rString, false)

	// We do need to cleanup, as otherwise we get some spurious changes on complex diffs
	diffs = dmp.DiffCleanupSemantic(diffs)

	var l, r string

	var results []lineRecord
	for _, diff := range diffs {
		lines := strings.Split(diff.Text, "\n")
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			if len(lines) > 0 {
				r += lines[0]
				if len(lines) > 1 {
					results = append(results, lineRecord{Type: diffmatchpatch.DiffInsert, Line: r})
					r = ""
				}
			}
			for i := 1; i < len(lines)-1; i++ {
				line := lines[i]
				results = append(results, lineRecord{Type: diffmatchpatch.DiffInsert, Line: line})
			}
			if len(lines) > 1 {
				r = lines[len(lines)-1]
			}

		case diffmatchpatch.DiffDelete:
			if len(lines) > 0 {
				l += lines[0]
				if len(lines) > 1 {
					results = append(results, lineRecord{Type: diffmatchpatch.DiffDelete, Line: l})
					l = ""
				}
			}
			for i := 1; i < len(lines)-1; i++ {
				line := lines[i]
				results = append(results, lineRecord{Type: diffmatchpatch.DiffDelete, Line: line})
			}
			if len(lines) > 1 {
				l = lines[len(lines)-1]
			}

		case diffmatchpatch.DiffEqual:
			if len(lines) == 1 {
				l += lines[0]
				r += lines[0]
			}
			if len(lines) > 1 {
				if l != "" || r != "" {
					l += lines[0]
					r += lines[0]
				} else {
					results = append(results, lineRecord{Type: diffmatchpatch.DiffEqual, Line: lines[0]})
				}
				if r != "" {
					results = append(results, lineRecord{Type: diffmatchpatch.DiffInsert, Line: r})
					r = ""
				}

				if l != "" {
					results = append(results, lineRecord{Type: diffmatchpatch.DiffDelete, Line: l})
					l = ""
				}

			}
			for i := 1; i < len(lines)-1; i++ {
				line := lines[i]
				results = append(results, lineRecord{Type: diffmatchpatch.DiffEqual, Line: line})
			}
			if len(lines) > 1 {
				l = lines[len(lines)-1]
				r = lines[len(lines)-1]
			}

		default:
			klog.Fatalf("unexpected dmp type: %v", diff.Type)
		}
	}

	if l != "" && l == r {
		results = append(results, lineRecord{Type: diffmatchpatch.DiffEqual, Line: l})
	} else {
		if l != "" {
			results = append(results, lineRecord{Type: diffmatchpatch.DiffDelete, Line: l})
		}
		if r != "" {
			results = append(results, lineRecord{Type: diffmatchpatch.DiffInsert, Line: r})
		}
	}

	return results
}
