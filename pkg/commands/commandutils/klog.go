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

package commandutils

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

// ConfigureKlogForCompletion configures Klog to not interfere with Cobra completion functions.
func ConfigureKlogForCompletion() {
	klog.SetOutput(&toCompDebug{})
	klog.LogToStderr(false)
}

type toCompDebug struct{}

func (t toCompDebug) Write(p []byte) (n int, err error) {
	cobra.CompDebug(string(p), false)
	return len(p), nil
}
