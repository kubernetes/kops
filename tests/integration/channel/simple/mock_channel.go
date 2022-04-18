/*
Copyright 2020 The Kubernetes Authors.

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

package simple

import (
	"fmt"
	"os"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/channel"
)

func NewMockChannel(sourcePath string) (*kops.Channel, error) {
	sourceBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("unexpected error reading sourcePath %q: %v", sourcePath, err)
	}

	channel, err := channel.ParseChannel(sourceBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse channel: %v", err)
	}
	return channel, nil
}
