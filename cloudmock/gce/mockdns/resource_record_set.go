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

package mockdns

import (
	dns "google.golang.org/api/dns/v1"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

type resourceRecordSetClient struct {
	// resourceRecordSets are resourceRecordSets keyed by project,zone and resourceRecordSet name.
	resourceRecordSets map[string]map[string]map[string]*dns.ResourceRecordSet
}

var _ gce.ResourceRecordSetClient = &resourceRecordSetClient{}

func newResourceRecordSetClient() *resourceRecordSetClient {
	return &resourceRecordSetClient{
		resourceRecordSets: map[string]map[string]map[string]*dns.ResourceRecordSet{},
	}
}

func (c *resourceRecordSetClient) List(project, zone string) ([]*dns.ResourceRecordSet, error) {
	zones, ok := c.resourceRecordSets[project]
	if !ok {
		return nil, nil
	}
	rs, ok := zones[zone]
	if !ok {
		return nil, nil
	}
	var l []*dns.ResourceRecordSet
	for _, r := range rs {
		l = append(l, r)
	}
	return l, nil
}
