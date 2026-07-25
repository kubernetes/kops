/*
Copyright 2026 The Kubernetes Authors.

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

package linodetasks

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"slices"
	"strings"

	"github.com/linode/linodego/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

// +kops:fitask
type Instance struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Region                 string
	Type                   string
	Image                  string
	Count                  int
	Subnet                 *Subnet
	RequirePublicInterface *bool
	Tags                   []string
	AuthorizedKeys         []*SSHKey
	UserData               fi.Resource
	NeedsUpdate            []string
}

var _ fi.CloudupTask = &Instance{}
var _ fi.CompareWithID = &Instance{}

// CompareWithID returns the name of the instance as its unique identifier.
func (i *Instance) CompareWithID() *string {
	return i.Name
}

// GetDependencies returns the dependencies of the instance, which is the subnet it belongs to.
func (i *Instance) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	if i.Subnet != nil {
		deps = append(deps, i.Subnet)
	}
	if i.AuthorizedKeys != nil {
		for _, key := range i.AuthorizedKeys {
			if key.PublicKey != nil {
				deps = append(deps, key)
			}
		}
	}
	return deps
}

func (i *Instance) Find(c *fi.CloudupContext) (*Instance, error) {
	cloud := c.T.Cloud.(linode.LinodeCloud)
	labelSelector := []string{
		fmt.Sprintf("%s:%s", linode.TagKubernetesClusterName, c.T.Cluster.Name),
		fmt.Sprintf("%s:%s", linode.TagKubernetesInstanceGroup, fi.ValueOf(i.Name)),
	}
	listOptions, err := linode.ListOptionsForTags(labelSelector...)
	if err != nil {
		return nil, err
	}

	instances, err := cloud.Client().ListInstances(c.Context(), listOptions)
	if err != nil {
		return nil, err
	}

	if len(instances) == 0 {
		return nil, nil
	}

	expectedTags := append([]string{}, i.Tags...)

	// Calculate the UserData hash to determine if the instance needs an update
	userDataBytes, err := fi.ResourceAsBytes(i.UserData)
	if err != nil {
		return nil, err
	}
	userDataHash := generateUserDataHash(string(userDataBytes))

	expectedTags = append(expectedTags, fmt.Sprintf("%s:%s", linode.TagKubernetesInstanceUserData, userDataHash))

	var needsUpdate []string
	for _, instance := range instances {
		if instance.Type != i.Type {
			needsUpdate = append(needsUpdate, instance.Label)
			continue
		}
		if instance.Image != i.Image {
			needsUpdate = append(needsUpdate, instance.Label)
			continue
		}
		if instance.Region != i.Region {
			needsUpdate = append(needsUpdate, instance.Label)
			continue
		}
		if !hasAllTags(instance.Tags, expectedTags) {
			needsUpdate = append(needsUpdate, instance.Label)
			continue
		}
		if i.Subnet == nil || i.Subnet.ID == nil {
			return nil, fmt.Errorf("Subnet.ID is required")
		}
		interfaces, err := cloud.Client().ListInterfaces(c.Context(), instance.ID, nil)
		if err != nil {
			return nil, fmt.Errorf("error listing Linode (Akamai) interfaces for instance %q: %w", instance.Label, err)
		}
		if !hasExpectedInterfaces(interfaces, fi.ValueOf(i.Subnet.ID), fi.ValueOf(i.RequirePublicInterface)) {
			needsUpdate = append(needsUpdate, instance.Label)
			continue
		}
	}

	actual := *i
	actual.NeedsUpdate = needsUpdate
	actual.Count = len(instances)

	return &actual, nil
}

func (i *Instance) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(i, c)
}

func (i *Instance) CheckChanges(actual, expected, changes *Instance) error {
	if actual != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.Region != "" {
			return fi.CannotChangeField("Region")
		}
		if changes.Subnet != nil {
			return fi.CannotChangeField("Subnet")
		}
		if changes.RequirePublicInterface != nil {
			return fi.CannotChangeField("RequirePublicInterface")
		}
	} else {
		if expected.Name == nil {
			return fi.RequiredField("Name")
		}
		if expected.Region == "" {
			return fi.RequiredField("Region")
		}
		if expected.Type == "" {
			return fi.RequiredField("Type")
		}
		if expected.Count < 0 {
			return fmt.Errorf("Count must be positive or 0")
		}
		if expected.Subnet == nil {
			return fi.RequiredField("Subnet")
		}
		if expected.RequirePublicInterface == nil {
			return fi.RequiredField("RequirePublicInterface")
		}
		if expected.Image == "" {
			return fi.RequiredField("Image")
		}
		if expected.UserData == nil {
			return fi.RequiredField("UserData")
		}
		if len(expected.AuthorizedKeys) < 1 {
			return fi.RequiredField("AuthorizedKeys")
		}
		if len(expected.Tags) == 0 {
			return fi.RequiredField("Tags")
		}
	}
	return nil
}

func (i *Instance) RenderLinode(t *linode.APITarget, actual, expected, changes *Instance) error {
	actualCount := 0
	if actual != nil {
		actualCount = actual.Count
	}
	if actualCount >= expected.Count {
		return nil
	}

	if expected.Subnet == nil || expected.Subnet.ID == nil {
		return fmt.Errorf("Subnet.ID is required")
	}

	authorizedKeys, err := resolveAuthorizedKeys(t.Cloud.Client(), expected.AuthorizedKeys)
	if err != nil {
		return err
	}

	userDataBytes, err := fi.ResourceAsBytes(expected.UserData)
	if err != nil {
		return err
	}
	userDataHash := generateUserDataHash(string(userDataBytes))
	instanceTags := append([]string{}, expected.Tags...)
	instanceTags = append(instanceTags, fmt.Sprintf("%s:%s", linode.TagKubernetesInstanceUserData, userDataHash))
	encodedUserData := base64.StdEncoding.EncodeToString(userDataBytes)

	interfaces := buildLinodeInterfaces(fi.ValueOf(expected.Subnet.ID), fi.ValueOf(expected.RequirePublicInterface))
	toCreate := expected.Count - actualCount
	for range toCreate {
		label, err := buildLinodeInstanceLabel(fi.ValueOf(expected.Name))
		if err != nil {
			return fmt.Errorf("error generating Linode (Akamai) instance label for group %q: %w", fi.ValueOf(expected.Name), err)
		}
		_, err = t.Cloud.Client().CreateInstance(context.Background(), linodego.InstanceCreateOptions{
			Region:              expected.Region,
			Type:                expected.Type,
			Label:               label,
			Image:               expected.Image,
			AuthorizedKeys:      authorizedKeys,
			Tags:                instanceTags,
			Metadata:            &linodego.InstanceMetadataOptions{UserData: encodedUserData},
			InterfaceGeneration: linodego.GenerationLinode,
			LinodeInterfaces:    interfaces,
		})
		if err != nil {
			return fmt.Errorf("error creating Linode (Akamai) instance for group %q: %w", fi.ValueOf(expected.Name), err)
		}
	}

	return nil
}

// generateUserDataHash generates a unique hash for the given user data string.
func generateUserDataHash(userData string) string {
	hash := sha256.Sum256([]byte(userData))
	fullHashString := base64.RawURLEncoding.EncodeToString(hash[:])

	// Truncate the hash to ensure the tag name does not exceed 50 characters.
	maxLength := 50 - len(linode.TagKubernetesInstanceUserData) - 1 // 1 for the colon separator
	truncatedHash := fullHashString[:maxLength]

	return truncatedHash
}

// resolveAuthorizedKeys resolves the authorized keys for the instance.
// It checks if the key has a public key directly provided or if it needs to be looked up by name in Linode (Akamai).
func resolveAuthorizedKeys(client linode.LinodeClient, keys []*SSHKey) ([]string, error) {
	var authorizedKeys []string
	var keysByName map[string]string

	for _, key := range keys {
		if key == nil {
			continue
		}
		if key.PublicKey != nil {
			publicKey, err := fi.ResourceAsString(*key.PublicKey)
			if err != nil {
				return nil, fmt.Errorf("error rendering SSH key data: %w", err)
			}
			authorizedKeys = append(authorizedKeys, strings.TrimSpace(publicKey))
			continue
		}

		if keysByName == nil {
			listedKeys, err := client.ListSSHKeys(context.TODO(), nil)
			if err != nil {
				return nil, fmt.Errorf("error listing Linode (Akamai) SSH keys: %w", err)
			}
			keysByName = make(map[string]string, len(listedKeys))
			for _, listedKey := range listedKeys {
				keysByName[listedKey.Label] = listedKey.SSHKey
			}
		}

		publicKey, found := keysByName[fi.ValueOf(key.Name)]
		if !found {
			return nil, fmt.Errorf("SSH key %q not found in Linode (Akamai)", fi.ValueOf(key.Name))
		}
		authorizedKeys = append(authorizedKeys, strings.TrimSpace(publicKey))
	}

	return authorizedKeys, nil
}

// buildLinodeInterfaces builds the Linode (Akamai) interfaces for the instance based on the subnet ID and whether a public interface is required.
func buildLinodeInterfaces(subnetID int, requirePublicInterface bool) []linodego.LinodeInterfaceCreateOptions {
	var interfaces []linodego.LinodeInterfaceCreateOptions
	if requirePublicInterface {
		interfaces = append(interfaces, linodego.LinodeInterfaceCreateOptions{
			Public: &linodego.PublicInterfaceCreateOptions{},
		})
	}
	interfaces = append(interfaces, linodego.LinodeInterfaceCreateOptions{
		VPC: &linodego.VPCInterfaceCreateOptions{SubnetID: subnetID},
	})

	return interfaces
}

// buildLinodeInstanceLabel generates a unique label for the Linode (Akamai) instance by appending a random suffix to the provided name.
// It ensures that the final label does not exceed 64 characters and trims any trailing hyphens, underscores, or periods.
func buildLinodeInstanceLabel(name string) (string, error) {
	var randomSuffix [8]byte
	if _, err := cryptorand.Read(randomSuffix[:]); err != nil {
		return "", fmt.Errorf("error generating random label suffix: %w", err)
	}

	suffix := fmt.Sprintf("-%x", randomSuffix)
	maxBaseLength := 64 - len(suffix)
	if len(name) > maxBaseLength {
		name = name[:maxBaseLength]
	}
	name = strings.TrimRight(name, "-_.")

	return name + suffix, nil
}

// hasExpectedInterfaces checks if the Linode (Akamai) instance has the expected network interfaces.
// It verifies that there is exactly one VPC interface with the matching subnet ID and checks for the presence of a public interface if required.
func hasExpectedInterfaces(interfaces []linodego.LinodeInterface, subnetID int, requirePublicInterface bool) bool {
	publicCount := 0
	vpcCount := 0
	hasMatchingVPCSubnet := false

	for _, iface := range interfaces {
		if iface.Public != nil {
			publicCount++
		}
		if iface.VPC != nil {
			vpcCount++
			if iface.VPC.SubnetID == subnetID {
				hasMatchingVPCSubnet = true
			}
		}
	}

	if vpcCount != 1 || !hasMatchingVPCSubnet {
		return false
	}

	if requirePublicInterface {
		return publicCount == 1
	}

	return publicCount == 0
}

// hasAllTags checks if all expected tags are present in the actual tags of the Linode (Akamai) instance.
func hasAllTags(actual, expected []string) bool {
	for _, tag := range expected {
		if !slices.Contains(actual, tag) {
			return false
		}
	}

	return true
}
