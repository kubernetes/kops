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

package awsup

import (
	"fmt"

	"github.com/aws/smithy-go/logging"
	"k8s.io/klog/v2"
)

type awsLogger struct{}

var _ logging.Logger = awsLogger{}

func (awsLogger) Logf(classification logging.Classification, format string, v ...interface{}) {
	text := fmt.Sprintf("AWS request: %s", format)
	switch classification {
	case logging.Warn:
		klog.Warningf(text, v...)
	default:
		klog.V(2).Infof(text, v...)
	}
}
