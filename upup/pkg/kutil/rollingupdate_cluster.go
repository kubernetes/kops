package kutil

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"sync"
	"time"
)

// RollingUpdateCluster restarts cluster nodes
type RollingUpdateCluster struct {
	ClusterName string
	Region      string
	Cloud       fi.Cloud
}

func (c *RollingUpdateCluster) ListNodesets() (map[string]*Nodeset, error) {
	cloud := c.Cloud.(*awsup.AWSCloud)

	nodesets := make(map[string]*Nodeset)

	tags := cloud.BuildTags(nil)

	asgs, err := findAutoscalingGroups(cloud, tags)
	if err != nil {
		return nil, err
	}

	for _, asg := range asgs {
		nodeset := buildNodeset(asg)
		nodesets[nodeset.Name] = nodeset
	}

	return nodesets, nil
}

func (c *RollingUpdateCluster) RollingUpdateNodesets(nodesets map[string]*Nodeset) error {
	if len(nodesets) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	var resultsMutex sync.Mutex
	results := make(map[string]error)

	for k, nodeset := range nodesets {
		wg.Add(1)
		go func(k string, nodeset *Nodeset) {
			resultsMutex.Lock()
			results[k] = fmt.Errorf("function panic")
			resultsMutex.Unlock()

			defer wg.Done()
			err := nodeset.RollingUpdate(c.Cloud)

			resultsMutex.Lock()
			results[k] = err
			resultsMutex.Unlock()
		}(k, nodeset)
	}

	wg.Wait()

	for _, err := range results {
		if err != nil {
			return err
		}
	}

	return nil
}

type Nodeset struct {
	Name       string
	Status     string
	Ready      []*autoscaling.Instance
	NeedUpdate []*autoscaling.Instance
}

func buildNodeset(g *autoscaling.Group) *Nodeset {
	n := &Nodeset{
		Name: aws.StringValue(g.AutoScalingGroupName),
	}

	findLaunchConfigurationName := aws.StringValue(g.LaunchConfigurationName)

	for _, i := range g.Instances {
		if findLaunchConfigurationName == aws.StringValue(i.LaunchConfigurationName) {
			n.Ready = append(n.Ready, i)
		} else {
			n.NeedUpdate = append(n.NeedUpdate, i)
		}
	}

	if len(n.NeedUpdate) == 0 {
		n.Status = "Ready"
	} else {
		n.Status = "NeedsUpdate"
	}

	return n
}

func (n *Nodeset) RollingUpdate(cloud fi.Cloud) error {
	c := cloud.(*awsup.AWSCloud)

	for _, i := range n.NeedUpdate {
		glog.Infof("Stopping instance %q in nodeset %q", *i.InstanceId, n.Name)

		// TODO: Evacuate through k8s first?

		// TODO: Temporarily increase size of ASG?

		// TODO: Remove from ASG first so status is immediately updated?

		// TODO: Batch termination, like a rolling-update

		request := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{i.InstanceId},
		}
		_, err := c.EC2.TerminateInstances(request)
		if err != nil {
			return fmt.Errorf("error deleting instance %q: %v", i.InstanceId, err)
		}

		// TODO: Wait for node to appear back in k8s
		time.Sleep(time.Minute)
	}

	return nil
}

func (n *Nodeset) String() string {
	return "nodeset:" + n.Name
}
