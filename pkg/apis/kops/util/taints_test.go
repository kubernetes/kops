/*
Copyright 2020 The Kubernetes Authors.

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

package util

import (
	"reflect"
	"testing"
)

func TestParseTaint(t *testing.T) {
	grid := []struct {
		name    string
		taint   string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "key only",
			taint: "key1",
			want: map[string]string{
				"key":    "key1",
				"value":  "",
				"effect": "",
			},
			wantErr: false,
		},
		{
			name:  "key with no value and effect",
			taint: "key1:NoSchedule",
			want: map[string]string{
				"key":    "key1",
				"value":  "",
				"effect": "NoSchedule",
			},
			wantErr: false,
		},
		{
			name:  "key with value and effect",
			taint: "key1=value1:NoSchedule",
			want: map[string]string{
				"key":    "key1",
				"value":  "value1",
				"effect": "NoSchedule",
			},
			wantErr: false,
		},
		{
			name:    "invalid taint spec (too many equals)",
			taint:   "key1=value1=value2:NoSchedule",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid taint spec (too many colons)",
			taint:   "key1:NoSchedule:NoExecute",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tc := range grid {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseTaint(tc.taint)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseTaint() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParseTaint() = %v, want %v", got, tc.want)
			}
		})
	}
}
