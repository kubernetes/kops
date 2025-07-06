/*
Copyright 2025 The Kubernetes Authors.

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

package elementotasks

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/elemento"
)

// +kops:fitask
type ServerGroup struct {
	Name      *string
	Lifecycle fi.Lifecycle
	SSHKeys   []*SSHKey
	Network   *Network

	Count      int
	NeedUpdate []string

	Location     string
	Size         string
	Image        string
	Architecture string

	EnableIPv4 bool
	EnableIPv6 bool

	UserData fi.Resource

	Labels map[string]string

	// RootVolumeSize is the size of the root volume in GB
	RootVolumeSize *int32
}

func (v *ServerGroup) Find(c *fi.CloudupContext) (*ServerGroup, error) {
	cloud := c.T.Cloud.(elemento.ElementoCloud)
	client := cloud.ServerClient()

	labelSelector := []string{
		fmt.Sprintf("%s=%s", elemento.TagKubernetesClusterName, c.T.Cluster.Name),
		fmt.Sprintf("%s=%s", elemento.TagKubernetesInstanceGroup, fi.ValueOf(v.Name)),
	}
	listOptions := ecloud.ListOpts{
		PerPage:       50,
		LabelSelector: strings.Join(labelSelector, ","),
	}
	serverListOptions := ecloud.ServerListOpts{ListOpts: listOptions}
	servers, err := client.AllWithOpts(context.TODO(), serverListOptions)
	if err != nil {
		return nil, err
	}
	if len(servers) == 0 {
		return nil, nil
	}

	// Calculate the user-data hash
	userDataBytes, err := fi.ResourceAsBytes(v.UserData)
	if err != nil {
		return nil, err
	}
	userDataHash := safeBytesHash(userDataBytes)

	// Add the expected user-data hash label
	v.Labels[elemento.TagKubernetesInstanceUserData] = userDataHash

	actual := *v
	actual.Count = len(servers)

	// Find servers that need to be updated
	for i, server := range servers {
		// Ignore servers that are already labeled as needing update
		if _, ok := server.Labels[elemento.TagKubernetesInstanceNeedsUpdate]; ok {
			continue
		}

		// Check if server index is higher than desired count
		if i >= v.Count {
			actual.NeedUpdate = append(actual.NeedUpdate, server.Name)
			continue
		}

		// Check if server matches the expected group template
		if server.Labels[elemento.TagKubernetesInstanceUserData] != userDataHash {
			actual.NeedUpdate = append(actual.NeedUpdate, server.Name)
			continue
		}
		if server.Datacenter.Location != v.Location {
			actual.NeedUpdate = append(actual.NeedUpdate, server.Name)
			continue
		}
		if server.ServerType == nil || server.ServerType.Name != v.Size {
			actual.NeedUpdate = append(actual.NeedUpdate, server.Name)
			continue
		}
		if (server.PublicNet.IPv4 != "") != v.EnableIPv4 {
			actual.NeedUpdate = append(actual.NeedUpdate, server.Name)
			continue
		}
		if (server.PublicNet.IPv6 != "") != v.EnableIPv6 {
			actual.NeedUpdate = append(actual.NeedUpdate, server.Name)
			continue
		}
		// TODO: Check root volume size when the ecloud library provides access to it
		// if server.RootVolumeSize != v.RootVolumeSize {
		//     actual.NeedUpdate = append(actual.NeedUpdate, server.Name)
		//     continue
		// }
	}

	return &actual, nil
}

func (v *ServerGroup) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(v, c)
}

func (*ServerGroup) CheckChanges(a, e, changes *ServerGroup) error {
	if e.Name == nil {
		return fi.RequiredField("Name")
	}
	if e.Location == "" {
		return fi.RequiredField("Location")
	}
	if e.Size == "" {
		return fi.RequiredField("Size")
	}
	if e.Image == "" {
		return fi.RequiredField("Image")
	}
	if e.UserData == nil {
		return fi.RequiredField("UserData")
	}
	if e.RootVolumeSize != nil && fi.ValueOf(e.RootVolumeSize) <= 0 {
		return fi.RequiredField("RootVolumeSize must be greater than 0")
	}
	return nil
}

func (*ServerGroup) RenderElemento(t *elemento.ElementoAPITarget, a, e, changes *ServerGroup) error {
	client := t.Cloud.ServerClient()

	if a != nil {
		// Add "kops.k8s.io/needs-update" label to servers needing update
		for _, serverName := range a.NeedUpdate {
			server, _, err := client.GetByName(context.TODO(), serverName)
			if err != nil {
				return err
			}
			if server == nil {
				continue
			}

			server.Labels[elemento.TagKubernetesInstanceNeedsUpdate] = ""
			_, _, err = client.Update(context.TODO(), server, ecloud.ServerUpdateOpts{
				Name:   server.Name,
				Labels: server.Labels,
			})
			if err != nil {
				return err
			}
		}
	}

	actualCount := 0
	if a != nil {
		actualCount = a.Count
	}
	expectedCount := e.Count

	if actualCount >= expectedCount {
		return nil
	}

	if len(e.SSHKeys) == 0 {
		return fmt.Errorf("failed to find ssh keys for server group %q", fi.ValueOf(e.Name))
	}
	if e.Network == nil {
		return fmt.Errorf("failed to find network for server group %q", fi.ValueOf(e.Name))
	}

	userData, err := fi.ResourceAsString(e.UserData)
	if err != nil {
		return err
	}
	userDataBytes, err := fi.ResourceAsBytes(e.UserData)
	if err != nil {
		return err
	}
	userDataHash := safeBytesHash(userDataBytes)

	for i := 1; i <= expectedCount-actualCount; i++ {
		// Append a random/unique ID to the node name
		name := fmt.Sprintf("%s-%x", fi.ValueOf(e.Name), rand.Int63())

		id, err := strconv.Atoi(fi.ValueOf(e.Network.ID))
		if err != nil {
			return err
		}
		opts := ecloud.ServerCreateOpts{
			Name:             name,
			StartAfterCreate: fi.PtrTo(true),
			Networks: []*ecloud.Network{
				{
					ID: id,
				},
			},
			Datacenter: &ecloud.Datacenter{
				Location: e.Location,
			},
			ServerType: &ecloud.ServerType{
				Name: e.Size,
			},
			UserData: userData,
			Labels:   e.Labels,
		}

		// Add root volume configuration if specified
		if e.RootVolumeSize != nil {
			opts.ServerType.Disk = int(fi.ValueOf(e.RootVolumeSize))
		}

		// Add the SSH keys
		for _, sshkey := range e.SSHKeys {
			opts.SSHKeys = append(opts.SSHKeys, &ecloud.SSHKey{ID: fi.ValueOf(sshkey.ID)})
		}

		// Add the user-data hash label
		opts.Labels[elemento.TagKubernetesInstanceUserData] = userDataHash

		_, _, err = client.Create(context.TODO(), opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func safeBytesHash(data []byte) string {
	// Calculate the SHA256 checksum of the data
	sum256 := sha256.Sum256(data)

	// Replace the unsupported chars with supported ones
	safe256 := base64.StdEncoding.EncodeToString(sum256[:])
	safe256 = strings.ReplaceAll(safe256, "+", "-")
	safe256 = strings.ReplaceAll(safe256, "/", "_")

	// Trim the unsupported "=" padding chars
	safe256 = strings.TrimRight(safe256, "=")

	return fmt.Sprintf("sha256.%s", safe256)
}
