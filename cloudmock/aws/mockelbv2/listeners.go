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

package mockelbv2

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"k8s.io/klog/v2"
)

func (m *MockELBV2) DescribeListeners(request *elbv2.DescribeListenersInput) (*elbv2.DescribeListenersOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeListeners v2 %v", request)

	resp := &elbv2.DescribeListenersOutput{
		Listeners: make([]*elbv2.Listener, 0),
	}
	for _, l := range m.Listeners {
		listener := l.description
		if aws.StringValue(request.LoadBalancerArn) == aws.StringValue(listener.LoadBalancerArn) {
			resp.Listeners = append(resp.Listeners, &listener)
		} else {
			for _, reqARN := range request.ListenerArns {
				if aws.StringValue(reqARN) == aws.StringValue(listener.ListenerArn) {
					resp.Listeners = append(resp.Listeners, &listener)
				}
			}
		}
	}
	return resp, nil
}

func (m *MockELBV2) CreateListener(request *elbv2.CreateListenerInput) (*elbv2.CreateListenerOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("CreateListener v2 %v", request)

	l := elbv2.Listener{
		DefaultActions:  request.DefaultActions,
		LoadBalancerArn: request.LoadBalancerArn,
		Port:            request.Port,
		Certificates:    request.Certificates,
		Protocol:        request.Protocol,
		SslPolicy:       request.SslPolicy,
	}

	lbARN := aws.StringValue(request.LoadBalancerArn)
	if _, ok := m.LoadBalancers[lbARN]; !ok {
		return nil, fmt.Errorf("LoadBalancerArn not found %v", aws.StringValue(request.LoadBalancerArn))
	}

	m.listenerCount++
	arn := fmt.Sprintf("%v/%v", strings.Replace(lbARN, ":loadbalancer/", ":listener/", 1), m.listenerCount)
	l.ListenerArn = aws.String(arn)

	if m.Listeners == nil {
		m.Listeners = make(map[string]*listener)
	}

	tgARN := aws.StringValue(l.DefaultActions[0].TargetGroupArn)

	if _, ok := m.TargetGroups[tgARN]; ok {
		found := false
		for _, lb := range m.TargetGroups[tgARN].description.LoadBalancerArns {
			if aws.StringValue(lb) == lbARN {
				found = true
				break
			}
		}
		if !found {
			m.TargetGroups[tgARN].description.LoadBalancerArns = append(m.TargetGroups[tgARN].description.LoadBalancerArns, aws.String(lbARN))
		}
	}

	m.Listeners[arn] = &listener{description: l}
	return &elbv2.CreateListenerOutput{Listeners: []*elbv2.Listener{&l}}, nil
}

func (m *MockELBV2) DeleteListener(request *elbv2.DeleteListenerInput) (*elbv2.DeleteListenerOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteListener v2 %v", request)

	lARN := aws.StringValue(request.ListenerArn)
	if _, ok := m.Listeners[lARN]; !ok {
		return nil, fmt.Errorf("Listener not found %v", lARN)
	}
	delete(m.Listeners, lARN)
	return nil, nil
}

func (m *MockELBV2) ModifyListener(request *elbv2.ModifyListenerInput) (*elbv2.ModifyListenerOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	klog.Fatalf("elbv2.ModifyListener() not implemented")
	return nil, nil
}
