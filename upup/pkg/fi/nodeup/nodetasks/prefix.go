/*
Copyright 2021 The Kubernetes Authors.

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

package nodetasks

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
)

type Prefix struct {
	Name string
}

var _ fi.HasName = &Prefix{}

func (f *Prefix) GetName() *string {
	return &f.Name
}

// String returns a string representation, implementing the Stringer interface
func (p *Prefix) String() string {
	return fmt.Sprintf("Prefix: %s", p.Name)
}

func (e *Prefix) Find(c *fi.Context) (*Prefix, error) {
	if c.Cluster.Spec.GetCloudProvider() != kops.CloudProviderAWS {
		return nil, fmt.Errorf("unsupported cloud provider: %s", c.Cluster.Spec.GetCloudProvider())
	}

	mac, err := getInstanceMetadataFirstValue("mac")
	if err != nil {
		return nil, err
	}

	prefixes, err := getInstanceMetadataList(path.Join("network/interfaces/macs/", mac, "/ipv6-prefix"))
	if err != nil {
		return nil, err
	}
	if len(prefixes) == 0 {
		return nil, nil
	}

	klog.V(2).Infof("found prefix for primary network interface: %q", prefixes[0])
	actual := &Prefix{
		Name: e.Name,
	}
	return actual, nil
}

func (e *Prefix) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Prefix) CheckChanges(a, e, changes *Prefix) error {
	return nil
}

func (_ *Prefix) RenderLocal(t *local.LocalTarget, a, e, changes *Prefix) error {
	mac, err := getInstanceMetadataFirstValue("mac")
	if err != nil {
		return err
	}

	interfaceId, err := getInstanceMetadataFirstValue(path.Join("network/interfaces/macs/", mac, "/interface-id"))
	if err != nil {
		return err
	}

	response, err := t.Cloud.(awsup.AWSCloud).EC2().AssignIpv6Addresses(&ec2.AssignIpv6AddressesInput{
		Ipv6PrefixCount:    fi.Int64(1),
		NetworkInterfaceId: fi.String(interfaceId),
	})
	if err != nil {
		return fmt.Errorf("failed to assign prefix: %w", err)
	}
	klog.V(2).Infof("assigned prefix to primary network interface: %q", fi.StringValue(response.AssignedIpv6Prefixes[0]))

	return nil
}

func getInstanceMetadataFirstValue(category string) (string, error) {
	values, err := getInstanceMetadataList(category)
	if err != nil {
		return "", err
	}
	if len(values) == 0 {
		return "", fmt.Errorf("failed to get %q from ec2 meta-data: not found", category)
	}

	return values[0], nil
}

func getInstanceMetadataList(category string) ([]string, error) {
	sess := session.Must(session.NewSession())
	metadata := ec2metadata.New(sess)
	linesStr, err := metadata.GetMetadata(category)
	if err != nil {
		var aerr awserr.RequestFailure
		if errors.As(err, &aerr) && aerr.StatusCode() == http.StatusNotFound {
			return nil, nil
		} else {
			return nil, fmt.Errorf("failed to get %q from ec2 meta-data: %v", category, err)
		}
	}

	var values []string
	for _, line := range strings.Split(linesStr, "\n") {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			values = append(values, line)
		}
	}

	return values, nil
}
