//
// Copyright (c) 2016 The heketi Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package heketitest

import (
	"testing"

	client "github.com/heketi/heketi/client/api/go-client"
	"github.com/heketi/tests"
)

func TestNewHeketiMockTestServer(t *testing.T) {
	c := &HeketiMockTestServerConfig{
		Auth:     true,
		AdminKey: "admin",
		UserKey:  "user",
		Logging:  true,
	}

	h := NewHeketiMockTestServer(c)
	tests.Assert(t, h != nil)
	tests.Assert(t, h.Ts != nil)
	tests.Assert(t, h.DbFile != "")
	tests.Assert(t, h.App != nil)
	h.Close()

	h = NewHeketiMockTestServerDefault()
	tests.Assert(t, h != nil)
	tests.Assert(t, h.Ts != nil)
	tests.Assert(t, h.DbFile != "")
	tests.Assert(t, h.App != nil)
}

func TestHeketiMockTestServer(t *testing.T) {
	c := &HeketiMockTestServerConfig{
		Auth:     true,
		AdminKey: "admin",
		UserKey:  "user",
	}

	h := NewHeketiMockTestServer(c)
	defer h.Close()

	api := client.NewClient(h.URL(), "admin", "admin")
	tests.Assert(t, api != nil)

	cluster, err := api.ClusterCreate()
	tests.Assert(t, err == nil)
	tests.Assert(t, cluster != nil)
	tests.Assert(t, len(cluster.Nodes) == 0)
	tests.Assert(t, len(cluster.Volumes) == 0)

	info, err := api.ClusterInfo(cluster.Id)
	tests.Assert(t, err == nil)
	tests.Assert(t, info.Id == cluster.Id)
	tests.Assert(t, len(info.Nodes) == 0)
	tests.Assert(t, len(info.Volumes) == 0)
}
