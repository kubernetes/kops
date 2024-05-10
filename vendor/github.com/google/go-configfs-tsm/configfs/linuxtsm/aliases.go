// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package linuxtsm

// The aliases.go file is for "convenience" functions when folks only want to use the
// Linux client.

import (
	"github.com/google/go-configfs-tsm/report"
	"go.uber.org/multierr"
)

// GetReport returns a one-shot configfs-tsm report given a report request.
func GetReport(req *report.Request) (*report.Response, error) {
	var err error
	client, err := MakeClient()
	if err != nil {
		return nil, err
	}
	r, err := report.Create(client, req)
	if err != nil {
		return nil, err
	}
	response, err := r.Get()
	return response, multierr.Combine(r.Destroy(), err)
}
