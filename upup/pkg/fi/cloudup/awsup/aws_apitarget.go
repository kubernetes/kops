package awsup

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"time"
)

type AWSAPITarget struct {
	Cloud *AWSCloud
}

var _ fi.Target = &AWSAPITarget{}

func NewAWSAPITarget(cloud *AWSCloud) *AWSAPITarget {
	return &AWSAPITarget{
		Cloud: cloud,
	}
}

func (t *AWSAPITarget) Finish(taskMap map[string]fi.Task) error {
	return nil
}

func (t *AWSAPITarget) AddAWSTags(id string, expected map[string]string) error {
	actual, err := t.Cloud.GetTags(id)
	if err != nil {
		return fmt.Errorf("unexpected error fetching tags for resource: %v", err)
	}

	missing := map[string]string{}
	for k, v := range expected {
		actualValue, found := actual[k]
		if found && actualValue == v {
			continue
		}
		missing[k] = v
	}

	if len(missing) != 0 {
		request := &ec2.CreateTagsInput{}
		request.Resources = []*string{&id}
		for k, v := range missing {
			request.Tags = append(request.Tags, &ec2.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}

		_, err := t.Cloud.EC2.CreateTags(request)
		if err != nil {
			return fmt.Errorf("error adding tags to resource %q: %v", id, err)
		}
	}

	return nil
}

func (t *AWSAPITarget) WaitForInstanceRunning(instanceID string) error {
	attempt := 0
	for {
		instance, err := t.Cloud.DescribeInstance(instanceID)
		if err != nil {
			return fmt.Errorf("error while waiting for instance to be running: %v", err)
		}

		if instance == nil {
			// TODO: Wait if we _just_ created the instance?
			return fmt.Errorf("instance not found while waiting for instance to be running")
		}

		state := "?"
		if instance.State != nil {
			state = aws.StringValue(instance.State.Name)
		}
		if state == "running" {
			return nil
		}
		glog.Infof("Waiting for instance %q to be running (current state is %q)", instanceID, state)

		time.Sleep(10 * time.Second)
		attempt++
		if attempt > 30 {
			return fmt.Errorf("timeout waiting for instance %q to be running, state was %q", instanceID, state)
		}
	}
}
