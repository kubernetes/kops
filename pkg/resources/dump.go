/*
Copyright 2017 The Kubernetes Authors.

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

package resources

import (
	"fmt"

	"github.com/golang/glog"
)

// Dumpable is the interface that Resources that can report into the dump should implement
type Dumpable interface {
	Dump(dump *Dump) error
}

// BuildDump gathers information about the cluster and returns an object for dumping
func BuildDump(resources map[string]*Resource) (*Dump, error) {
	dump := &Dump{}

	for k, r := range resources {
		if r.Dumper == nil {
			glog.V(8).Infof("skipping dump of %q (does not implement Dumpable)", k)
			continue
		}

		err := r.Dumper(r, dump)
		if err != nil {
			return nil, fmt.Errorf("error dumping %q: %v", k, err)
		}
	}

	return dump, nil
}
