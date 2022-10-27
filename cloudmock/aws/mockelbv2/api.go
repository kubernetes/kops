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

package mockelbv2

import (
	"sync"

	"k8s.io/kops/cloudmock/aws/mockec2"

	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
)

type MockELBV2 struct {
	elbv2iface.ELBV2API

	mutex sync.Mutex

	EC2           *mockec2.MockEC2
	LoadBalancers map[string]*loadBalancer
	lbCount       int
	TargetGroups  map[string]*targetGroup
	tgCount       int
	Listeners     map[string]*listener
	listenerCount int
	LBAttributes  map[string][]*elbv2.LoadBalancerAttribute

	Tags map[string]*elbv2.TagDescription
}

type loadBalancer struct {
	description elbv2.LoadBalancer
}

type targetGroup struct {
	description elbv2.TargetGroup
}

type listener struct {
	description elbv2.Listener
}
