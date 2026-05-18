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
	"encoding/base64"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/linode/linodego"
	"k8s.io/kops/pkg/truncate"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

// +kops:fitask
type Instance struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Region        *string
	Type          *string
	Image         *string
	Count         int
	Tags          []string
	AuthorizedKey *fi.Resource
	UserData      *fi.Resource
}

var _ fi.CloudupTask = &Instance{}
var _ fi.CompareWithID = &Instance{}

var invalidInstanceLabelChars = regexp.MustCompile(`[^a-z0-9._-]+`)

func (i *Instance) CompareWithID() *string {
	return i.Name
}

func (i *Instance) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask

	if i.UserData != nil {
		deps = append(deps, fi.FindDependencies(tasks, i.UserData)...)
	}

	return deps
}

func (i *Instance) Find(c *fi.CloudupContext) (*Instance, error) {
	cloud := c.T.Cloud.(linode.LinodeCloud)

	instances, err := cloud.Client().ListInstances(c.Context(), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) instances: %w", err)
	}

	var matched []linodego.Instance
	for idx := range instances {
		instance := instances[idx]
		if !hasAllTags(instance.Tags, i.Tags) {
			continue
		}
		matched = append(matched, instance)
	}

	if len(matched) == 0 {
		return nil, nil
	}

	first := matched[0]
	return &Instance{
		Name:          i.Name,
		Lifecycle:     i.Lifecycle,
		Region:        fi.PtrTo(first.Region),
		Type:          fi.PtrTo(first.Type),
		Image:         fi.PtrTo(first.Image),
		Count:         len(matched),
		Tags:          append([]string(nil), first.Tags...),
		AuthorizedKey: i.AuthorizedKey,
		UserData:      i.UserData,
	}, nil
}

func (i *Instance) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(i, c)
}

func (_ *Instance) CheckChanges(actual, expected, changes *Instance) error {
	if expected.Count < 0 {
		return fmt.Errorf("Count cannot be negative")
	}

	if actual != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.Region != nil {
			return fi.CannotChangeField("Region")
		}
		if changes.Type != nil {
			return fi.CannotChangeField("Type")
		}
		if changes.Image != nil {
			return fi.CannotChangeField("Image")
		}
	} else {
		if expected.Name == nil {
			return fi.RequiredField("Name")
		}
		if expected.Region == nil {
			return fi.RequiredField("Region")
		}
		if expected.Type == nil {
			return fi.RequiredField("Type")
		}
		if expected.Image == nil {
			return fi.RequiredField("Image")
		}
	}

	return nil
}

func (_ *Instance) RenderLinode(t *linode.APITarget, actual, expected, changes *Instance) error {
	desiredCount := expected.Count
	actualCount := 0
	if actual != nil {
		actualCount = actual.Count
	}

	if desiredCount < actualCount {
		return fmt.Errorf("decreasing Linode (Akamai) instance count from %d to %d is not supported yet", actualCount, desiredCount)
	}
	if desiredCount == actualCount {
		return nil
	}

	encodedUserData, err := encodeUserData(expected.UserData)
	if err != nil {
		return err
	}

	var authorizedKeys []string
	if expected.AuthorizedKey != nil {
		publicKey, err := fi.ResourceAsString(*expected.AuthorizedKey)
		if err != nil {
			return fmt.Errorf("error rendering SSH public key: %w", err)
		}
		if trimmed := strings.TrimSpace(publicKey); trimmed != "" {
			authorizedKeys = append(authorizedKeys, trimmed)
		}
	}

	for ordinal := actualCount + 1; ordinal <= desiredCount; ordinal++ {
		rootPass, err := generateRootPassword()
		if err != nil {
			return err
		}

		opts := linodego.InstanceCreateOptions{
			Region:         fi.ValueOf(expected.Region),
			Type:           fi.ValueOf(expected.Type),
			Label:          makeInstanceLabel(fi.ValueOf(expected.Name), ordinal),
			RootPass:       rootPass,
			AuthorizedKeys: authorizedKeys,
			Image:          fi.ValueOf(expected.Image),
			PrivateIP:      true,
			Tags:           expected.Tags,
		}
		if encodedUserData != "" {
			opts.Metadata = &linodego.InstanceMetadataOptions{UserData: encodedUserData}
		}

		if _, err := t.Cloud.Client().CreateInstance(context.Background(), opts); err != nil {
			return fmt.Errorf("error creating Linode (Akamai) instance %q: %w", opts.Label, err)
		}
	}

	return nil
}

func hasAllTags(actual, expected []string) bool {
	for _, tag := range expected {
		if !slices.Contains(actual, tag) {
			return false
		}
	}
	return true
}

func encodeUserData(userData *fi.Resource) (string, error) {
	if userData == nil {
		return "", nil
	}
	bytes, err := fi.ResourceAsBytes(*userData)
	if err != nil {
		return "", fmt.Errorf("error rendering user-data: %w", err)
	}
	if len(bytes) == 0 {
		return "", nil
	}

	return base64.StdEncoding.EncodeToString(bytes), nil
}

func makeInstanceLabel(base string, ordinal int) string {
	clean := sanitizeLabelPart(base)
	if clean == "" {
		clean = "kops-node"
	}

	suffix := fmt.Sprintf("-%d", ordinal)
	maxBaseLength := max(64-len(suffix), 8)

	clean = truncate.TruncateString(clean, truncate.TruncateStringOptions{MaxLength: maxBaseLength})
	clean = strings.Trim(clean, "-_.")
	if clean == "" {
		clean = "kops-node"
	}

	// Insert ordinal before the first dot to ensure proper DNS zone matching
	// e.g., "control-plane-us-ord.masters.cluster.example.com" becomes
	// "control-plane-us-ord-1.masters.cluster.example.com" instead of
	// "control-plane-us-ord.masters.cluster.example.com-1"
	dotIndex := strings.IndexByte(clean, '.')
	if dotIndex != -1 {
		return clean[:dotIndex] + suffix + clean[dotIndex:]
	}

	return clean + suffix
}

func sanitizeLabel(s string, re *regexp.Regexp, trimSet string) string {
	s = strings.ToLower(s)
	s = re.ReplaceAllString(s, "-")
	s = collapseAdjacentSeparators(s)
	s = strings.Trim(s, trimSet)
	return s
}

func sanitizeLabelPart(s string) string {
	return sanitizeLabel(s, invalidInstanceLabelChars, "-_.")
}

func collapseAdjacentSeparators(s string) string {
	if s == "" {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))

	var previous rune
	for _, r := range s {
		if (r == '-' || r == '_' || r == '.') && r == previous {
			continue
		}
		b.WriteRune(r)
		previous = r
	}

	return b.String()
}

func generateRootPassword() (string, error) {
	secret, err := fi.CreateSecret()
	if err != nil {
		return "", fmt.Errorf("error generating root password: %w", err)
	}
	return string(secret.Data), nil
}
