package awsup

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"os"
)

// ValidateRegion checks that an AWS region name is valid
func ValidateRegion(region string) error {
	glog.V(2).Infof("Querying EC2 for all valid regions")

	request := &ec2.DescribeRegionsInput{}
	config := aws.NewConfig().WithRegion("us-east-1")
	client := ec2.New(session.New(), config)

	response, err := client.DescribeRegions(request)

	if err != nil {
		return fmt.Errorf("Got an error while querying for valid regions (verify your AWS credentials?)")
	}
	for _, r := range response.Regions {
		name := aws.StringValue(r.RegionName)
		if name == region {
			return nil
		}
	}

	if os.Getenv("SKIP_REGION_CHECK") != "" {
		glog.Infof("AWS region does not appear to be valid, but skipping because SKIP_REGION_CHECK is set")
		return nil
	}

	return fmt.Errorf("Region is not a recognized EC2 region: %q (check you have specified valid zones?)", region)
}
