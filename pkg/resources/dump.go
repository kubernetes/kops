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
	"context"
	"fmt"
	"sort"

	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
)

// DumpOperation holds context information for a dump, allowing for extension
type DumpOperation struct {
	// Context is the golang context.Context for the dump operation
	Context context.Context

	// Cloud is the cloud we are dumping
	Cloud fi.Cloud

	// CloudState allows the cloudprovider to store state during the dump operation
	CloudState interface{}

	// Dump is the target of our dump
	Dump *Dump
}

// BuildDump gathers information about the cluster and returns an object for dumping
func BuildDump(ctx context.Context, cloud fi.Cloud, resources map[string]*Resource) (*Dump, error) {
	dump := &Dump{}
	op := &DumpOperation{
		Context: ctx,
		Cloud:   cloud,
		Dump:    dump,
	}

	for k, r := range resources {
		if r.Dumper == nil {
			klog.V(8).Infof("skipping dump of %q (does not implement Dumpable)", k)
			continue
		}

		err := r.Dumper(op, r)
		if err != nil {
			return nil, fmt.Errorf("error dumping %q: %v", k, err)
		}
	}

	sort.SliceStable(dump.Instances, func(i, j int) bool { return dump.Instances[i].Name < dump.Instances[j].Name })

	return dump, nil
}
