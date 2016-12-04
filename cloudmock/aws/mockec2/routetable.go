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

package mockec2

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
)

func (m *MockEC2) DescribeRouteTablesRequest(*ec2.DescribeRouteTablesInput) (*request.Request, *ec2.DescribeRouteTablesOutput) {
	panic("Not implemented")
	return nil, nil
}

func (m *MockEC2) DescribeRouteTables(request *ec2.DescribeRouteTablesInput) (*ec2.DescribeRouteTablesOutput, error) {
	if request.Filters != nil {
		glog.Fatalf("filters not implemented: %v", request.Filters)
	}
	if request.DryRun != nil {
		glog.Fatalf("DryRun not implemented")
	}
	if request.RouteTableIds != nil {
		glog.Fatalf("RouteTableIds not implemented")
	}

	response := &ec2.DescribeRouteTablesOutput{}
	for _, rt := range m.RouteTables {
		response.RouteTables = append(response.RouteTables, rt)
	}
	return response, nil
}
