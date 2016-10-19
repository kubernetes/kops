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
