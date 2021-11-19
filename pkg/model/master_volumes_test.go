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

package model

import (
	"testing"
)

func TestValidateAWSVolumeAllow50ratio(t *testing.T) {
	volumeName := "a"
	volumeType := "io1"
	volumeIops := 1000
	volumeThroughput := 0
	volumeSize := 20

	err := validateAWSVolume(volumeName, volumeType, int32(volumeSize), int32(volumeIops), int32(volumeThroughput))
	if err != nil {
		t.Errorf("Failed to validate valid etcd member spec: %v", err)
	}
}
