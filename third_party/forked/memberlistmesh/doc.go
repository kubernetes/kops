// Copyright 2018 Prometheus Team
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

// Package cluster is an in-tree copy of github.com/jacksontj/memberlistmesh
// at commit 93462b9d2bb7 (pseudo-version
// v0.0.0-20190905163944-93462b9d2bb7). Upstream is no longer maintained.
//
// Modifications relative to upstream:
//
//   - protobuf runtime switched from github.com/gogo/protobuf/proto to
//     google.golang.org/protobuf/proto; clusterpb/cluster.pb.go is regenerated
//     with protoc-gen-go, which changes FullState.Parts from []Part to []*Part
//   - klog imports updated from k8s.io/klog to k8s.io/klog/v2
//   - github.com/pkg/errors replaced with stdlib errors and fmt.Errorf("%w")
//   - ulid entropy source simplified to crypto/rand.Reader
//   - interface{} replaced with any
//   - a handful of go vet / staticcheck cleanups (non-constant format strings,
//     error-string capitalization)
package cluster
