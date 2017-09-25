/*
Copyright 2016 The Kubernetes Authors.

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

package shared

import (
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

// AddSecurityGroups creates an slice of security groups from a slice of strings that container the security group id.
// If the context is defined, the task is ensured.
func AddSecurityGroups(groupIds []string, lifecycle *fi.Lifecycle, linkToVPC *awstasks.VPC, c *fi.ModelBuilderContext, link *awstasks.SecurityGroup) []*awstasks.SecurityGroup {
	var groups []*awstasks.SecurityGroup

	for i, id := range groupIds {
		name := fi.String(id)
		if link != nil && i == 0 {
			name = link.Name
		}
		t := &awstasks.SecurityGroup{
			Name:      name,
			ID:        fi.String(id),
			Shared:    fi.Bool(true),
			Lifecycle: lifecycle,

			VPC: linkToVPC,
		}

		groups = append(groups, t)

		if c != nil {
			c.EnsureTask(t)
		} else {
			glog.V(8).Infof("not ensuring security group: %q is in context", id)
		}
	}

	return groups
}
