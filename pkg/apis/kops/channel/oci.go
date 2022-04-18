/*
Copyright 2022 The Kubernetes Authors.

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

package channel

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/google/go-containerregistry/pkg/crane"
	"k8s.io/klog/v2"
	api "k8s.io/kops/pkg/apis/kops"
)

// isOCI returns true if the URL refers to an OCI image (container image)
func isOCI(u *url.URL) bool {
	return u.Scheme == "oci"
}

func resolveOCIChannel(u *url.URL) (*url.URL, error) {
	imageSpec, err := parseOCIImageSpec(u)
	if err != nil {
		return nil, err
	}

	if imageSpec.Tag != "" {
		// We already have a tag
		return u, nil
	}

	// Find the newest tag
	var options []crane.Option
	tags, err := crane.ListTags(imageSpec.Name, options...)
	if err != nil {
		return nil, fmt.Errorf("error listing tags for %s: %w", imageSpec.Name, err)
	}

	var latestTag string
	var latestVersion semver.Version
	for _, tag := range tags {
		version, err := semver.ParseTolerant(tag)
		if err != nil {
			klog.Warningf("ignoring unparseable tag %q", tag)
		}
		if latestTag == "" || version.Compare(latestVersion) >= 0 {
			latestTag = tag
			latestVersion = version
		}
	}

	if latestTag == "" {
		return nil, fmt.Errorf("could not find any tags for %s", u.String())
	}

	klog.Warningf("found latest tag %q for %v", latestTag, u.String())
	withVersion := &url.URL{}
	*withVersion = *u
	withVersion.Path += "@" + latestTag
	return withVersion, nil
}

type ociImageEntry struct {
	Path     string
	Header   tar.Header
	Contents []byte
}

type ociImage struct {
	Entries map[string]*ociImageEntry
}

type ociImageSpec struct {
	Name string
	Tag  string
}

func parseOCIImageSpec(u *url.URL) (*ociImageSpec, error) {
	if !isOCI(u) {
		return nil, fmt.Errorf("expected scheme to be oci in %q", u.String())
	}

	spec := &ociImageSpec{}

	pathTokens := strings.Split(u.Path, "@")
	if len(pathTokens) == 1 {
		// No tag
	} else if len(pathTokens) == 2 {
		spec.Tag = pathTokens[1]
	} else {
		return nil, fmt.Errorf("found multiple @ versions in %q", u.String())
	}

	path := pathTokens[0]
	path = strings.Trim(path, "/")

	spec.Name = u.Host + "/" + path

	return spec, nil
}

func loadOCIImage(resolvedURL *url.URL) (*ociImage, error) {
	resolved := resolvedURL.String()

	imageSpec, err := parseOCIImageSpec(resolvedURL)
	if err != nil {
		return nil, err
	}
	klog.V(2).Infof("Loading channel from %q", resolved)

	if imageSpec.Tag == "" {
		return nil, fmt.Errorf("expected @ version in %q", resolved)
	}

	srcImage := imageSpec.Name + ":" + imageSpec.Tag
	var pullOptions []crane.Option
	img, err := crane.Pull(srcImage, pullOptions...)
	if err != nil {
		return nil, fmt.Errorf("error pulling %s: %w", srcImage, err)
	}

	// We don't expect these to be very large, so we expand them entirely in-memory
	// (without e.g. implementing streaming)
	var b bytes.Buffer
	if err := crane.Export(img, &b); err != nil {
		return nil, fmt.Errorf("error exporting %s: %w", srcImage, err)
	}

	tarFile := &ociImage{
		Entries: make(map[string]*ociImageEntry),
	}
	tr := tar.NewReader(&b)
	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading tar header: %w", err)
		}

		b, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("error reading tar contents for %q: %w", header.Name, err)
		}
		tarFile.Entries[header.Name] = &ociImageEntry{
			Path:     header.Name,
			Header:   *header,
			Contents: b,
		}
	}

	return tarFile, nil
}

func LoadChannelFromOCI(resolvedURL *url.URL) (*api.Channel, error) {
	image, err := loadOCIImage(resolvedURL)
	if err != nil {
		return nil, err
	}

	entry := image.Entries["channel.yaml"]
	if entry == nil {
		return nil, fmt.Errorf("channel.yaml entry not found in %s", resolvedURL.String())
	}

	channel, err := ParseChannel(entry.Contents)
	if err != nil {
		return nil, fmt.Errorf("error parsing channel %q: %v", resolvedURL.String(), err)
	}
	klog.V(4).Infof("Channel contents: %s", string(entry.Contents))

	return channel, nil
}
