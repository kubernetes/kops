package awstasks

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kube-deploy/upup/pkg/fi"
)

type BlockDeviceMapping struct {
	VirtualName *string
}

func BlockDeviceMappingFromEC2(i *ec2.BlockDeviceMapping) (string, *BlockDeviceMapping) {
	o := &BlockDeviceMapping{}
	o.VirtualName = i.VirtualName
	return aws.StringValue(i.DeviceName), o
}

func (i *BlockDeviceMapping) ToEC2(deviceName string) *ec2.BlockDeviceMapping {
	o := &ec2.BlockDeviceMapping{}
	o.DeviceName = aws.String(deviceName)
	o.VirtualName = i.VirtualName
	return o
}

func BlockDeviceMappingFromAutoscaling(i *autoscaling.BlockDeviceMapping) (string, *BlockDeviceMapping) {
	o := &BlockDeviceMapping{}
	o.VirtualName = i.VirtualName
	return aws.StringValue(i.DeviceName), o
}

func (i *BlockDeviceMapping) ToAutoscaling(deviceName string) *autoscaling.BlockDeviceMapping {
	o := &autoscaling.BlockDeviceMapping{}
	o.DeviceName = aws.String(deviceName)
	o.VirtualName = i.VirtualName
	return o
}

var _ fi.HasDependencies = &BlockDeviceMapping{}

func (f *BlockDeviceMapping) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}
