/*
Copyright 2024 The Kubernetes Authors.

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

package metal

import (
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

type APITarget struct {
	Cloud       *Cloud
	OtherClouds []fi.Cloud
}

var _ fi.CloudupTarget = &APITarget{}

func NewAPITarget(cloud *Cloud, otherClouds []fi.Cloud) *APITarget {
	return &APITarget{
		Cloud:       cloud,
		OtherClouds: otherClouds,
	}
}

func (t *APITarget) GetAWSCloud() awsup.AWSCloud {
	klog.Fatalf("cannot find instance of AWSCloud in context")
	return nil
}

func (t *APITarget) Finish(taskMap map[string]fi.CloudupTask) error {
	return nil
}

func (t *APITarget) DefaultCheckExisting() bool {
	return true
}
