/*
Copyright 2023 The Kubernetes Authors.

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

package azuretasks

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-05-01/network"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

// ApplicationSecurityGroup is an Azure Cloud Application Security Group
// +kops:fitask
type ApplicationSecurityGroup struct {
	Name          *string
	ID            *string
	Lifecycle     fi.Lifecycle
	ResourceGroup *ResourceGroup

	Tags map[string]*string
}

var (
	_ fi.CloudupTask          = &ApplicationSecurityGroup{}
	_ fi.CompareWithID        = &ApplicationSecurityGroup{}
	_ fi.CloudupTaskNormalize = &ApplicationSecurityGroup{}
)

// CompareWithID returns the Name of the Application Security Group
func (asg *ApplicationSecurityGroup) CompareWithID() *string {
	return asg.ID
}

// Find discovers the Application Security Group in the cloud provider
func (asg *ApplicationSecurityGroup) Find(c *fi.CloudupContext) (*ApplicationSecurityGroup, error) {
	cloud := c.T.Cloud.(azure.AzureCloud)
	l, err := cloud.ApplicationSecurityGroup().List(context.TODO(), *asg.ResourceGroup.Name)
	if err != nil {
		return nil, err
	}
	var found *network.ApplicationSecurityGroup
	for _, v := range l {
		if *v.Name == *asg.Name {
			found = &v
			break
		}
	}
	if found == nil {
		return nil, nil
	}

	asg.ID = found.ID

	return &ApplicationSecurityGroup{
		Name:      asg.Name,
		Lifecycle: asg.Lifecycle,
		ResourceGroup: &ResourceGroup{
			Name: asg.ResourceGroup.Name,
		},
		ID:   found.ID,
		Tags: found.Tags,
	}, nil
}

func (asg *ApplicationSecurityGroup) Normalize(c *fi.CloudupContext) error {
	c.T.Cloud.(azure.AzureCloud).AddClusterTags(asg.Tags)
	return nil
}

// Run implements fi.Task.Run.
func (asg *ApplicationSecurityGroup) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(asg, c)
}

// CheckChanges returns an error if a change is not allowed.
func (*ApplicationSecurityGroup) CheckChanges(a, e, changes *ApplicationSecurityGroup) error {
	if a == nil {
		// Check if required fields are set when a new resource is created.
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		return nil
	}

	// Check if unchangeable fields won't be changed.
	if changes.Name != nil {
		return fi.CannotChangeField("Name")
	}
	return nil
}

// RenderAzure creates or updates a Application Security Group.
func (*ApplicationSecurityGroup) RenderAzure(t *azure.AzureAPITarget, a, e, changes *ApplicationSecurityGroup) error {
	if a == nil {
		klog.Infof("Creating a new Application Security Group with name: %s", fi.ValueOf(e.Name))
	} else {
		klog.Infof("Updating a Application Security Group with name: %s", fi.ValueOf(e.Name))
	}

	p := network.ApplicationSecurityGroup{
		Location: to.StringPtr(t.Cloud.Region()),
		Name:     to.StringPtr(*e.Name),
		Tags:     e.Tags,
	}

	asg, err := t.Cloud.ApplicationSecurityGroup().CreateOrUpdate(
		context.TODO(),
		*e.ResourceGroup.Name,
		*e.Name,
		p)
	if err != nil {
		return err
	}

	e.ID = asg.ID

	return nil
}
