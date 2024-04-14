package awsapi

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type SelectorInterface interface {
	ec2.DescribeInstanceTypeOfferingsAPIClient
	ec2.DescribeInstanceTypesAPIClient
	DescribeAvailabilityZones(ctx context.Context, params *ec2.DescribeAvailabilityZonesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeAvailabilityZonesOutput, error)
}
