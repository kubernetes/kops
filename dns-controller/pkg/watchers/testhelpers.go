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

package watchers

import "k8s.io/kops/dns-controller/pkg/dns"

type fakeScope struct {
	readyCh chan struct{}
	records map[string][]dns.Record
}

func (f *fakeScope) Replace(recordName string, records []dns.Record) {
	f.records[recordName] = records
}

// MarkReady is called when a watcher has processed all initial resources
//
// Since we don't care about watching for further updates, we just break of the go routine
func (f *fakeScope) MarkReady() {
	close(f.readyCh)
}

func (*fakeScope) AllKeys() []string {
	return []string{}
}

type fakeDNSContext struct {
	scope *fakeScope
}

func (f *fakeDNSContext) CreateScope(name string) (dns.Scope, error) {
	return f.scope, nil
}
