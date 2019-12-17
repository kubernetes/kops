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

package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// HasOwnedTag looks for the new tag indicating that the cluster does owns the resource, or the legacy tag
func HasOwnedTag(description string, tags []*ec2.Tag, clusterName string) bool {
	tagKey := "kubernetes.io/cluster/" + clusterName

	var found *ec2.Tag
	for _, tag := range tags {
		if aws.StringValue(tag.Key) != tagKey {
			continue
		}

		found = tag
	}

	if found != nil {
		tagValue := aws.StringValue(found.Value)
		switch tagValue {
		case "owned":
			return true
		case "shared":
			return false

		default:
			klog.Warningf("unknown cluster tag on %s: %q=%q", description, tagKey, tagValue)
			return false
		}
	}

	// Look for legacy tag - we assume that implies ownership
	for _, tag := range tags {
		if aws.StringValue(tag.Key) != awsup.TagClusterName {
			continue
		}

		found = tag
	}

	if found != nil {
		return true
	}

	// We warn here, because we shouldn't have found the object other than via a tag
	klog.Warningf("cluster tag not found on %s", description)
	return false
}
