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

package awsup

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"k8s.io/klog"
)

// RequestLogger logs every AWS request
type RequestLogger struct {
	logLevel klog.Level
}

func newRequestLogger(logLevel int) func(r *request.Request) {
	rl := &RequestLogger{
		logLevel: klog.Level(logLevel),
	}
	return rl.log
}

// Handler for aws-sdk-go that logs all requests
func (l *RequestLogger) log(r *request.Request) {
	service := r.ClientInfo.ServiceName
	name := "?"
	if r.Operation != nil {
		name = r.Operation.Name
	}
	methodDescription := service + "/" + name

	klog.V(l.logLevel).Infof("AWS request: %s", methodDescription)
}
