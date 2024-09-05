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
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	smithyhttp "github.com/aws/smithy-go/transport/http"
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

func (e *Prefix) Find(c *fi.NodeupContext) (*Prefix, error) {
	if c.T.BootConfig.CloudProvider != kops.CloudProviderAWS {
		return nil, fmt.Errorf("unsupported cloud provider: %s", c.T.BootConfig.CloudProvider)
	}

	mac, err := getInstanceMetadataFirstValue(c.Context(), "mac")
	if err != nil {
		return nil, err
	}

	prefixes, err := getInstanceMetadataList(c.Context(), path.Join("network/interfaces/macs/", mac, "/ipv6-prefix"))
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

func (e *Prefix) Run(c *fi.NodeupContext) error {
	return fi.NodeupDefaultDeltaRunMethod(e, c)
}

func (_ *Prefix) CheckChanges(a, e, changes *Prefix) error {
	return nil
}

func (_ *Prefix) RenderLocal(t *local.LocalTarget, a, e, changes *Prefix) error {
	ctx := context.TODO()
	mac, err := getInstanceMetadataFirstValue(ctx, "mac")
	if err != nil {
		return err
	}

	interfaceId, err := getInstanceMetadataFirstValue(ctx, path.Join("network/interfaces/macs/", mac, "/interface-id"))
	if err != nil {
		return err
	}

	response, err := t.Cloud.(awsup.AWSCloud).EC2().AssignIpv6Addresses(ctx, &ec2.AssignIpv6AddressesInput{
		Ipv6PrefixCount:    fi.PtrTo(int32(1)),
		NetworkInterfaceId: fi.PtrTo(interfaceId),
	})
	if err != nil {
		return fmt.Errorf("failed to assign prefix: %w", err)
	}
	klog.V(2).Infof("assigned prefix to primary network interface: %q", response.AssignedIpv6Prefixes[0])

	return nil
}

func getInstanceMetadataFirstValue(ctx context.Context, category string) (string, error) {
	values, err := getInstanceMetadataList(ctx, category)
	if err != nil {
		return "", err
	}
	if len(values) == 0 {
		return "", fmt.Errorf("failed to get %q from ec2 meta-data: not found", category)
	}

	return values[0], nil
}

func getInstanceMetadataList(ctx context.Context, category string) ([]string, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %v", err)
	}
	metadata := imds.NewFromConfig(cfg)
	resp, err := metadata.GetMetadata(ctx, &imds.GetMetadataInput{Path: category})
	if err != nil {
		var awsErr *smithyhttp.ResponseError
		if errors.As(err, &awsErr) && awsErr.HTTPStatusCode() == http.StatusNotFound {
			return nil, nil
		} else {
			return nil, fmt.Errorf("failed to get %q from ec2 meta-data: %v", category, err)
		}
	}
	defer resp.Content.Close()
	lines, err := io.ReadAll(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q from ec2 meta-data: %v", category, err)
	}

	var values []string
	for _, line := range strings.Split(string(lines), "\n") {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			values = append(values, line)
		}
	}

	return values, nil
}
