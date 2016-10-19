package mockec2

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type MockEC2 struct {
	RouteTables []*ec2.RouteTable
}

var _ ec2iface.EC2API = &MockEC2{}
