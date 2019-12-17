/*
Copyright 2017 The Kubernetes Authors.

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

package assettasks

import (
	"fmt"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
	"k8s.io/klog"
)

// dockerAPI encapsulates access to docker via the API
type dockerAPI struct {
	client *client.Client
}

// newDockerAPI builds a dockerAPI object, for talking to docker via the API
func newDockerAPI() (*dockerAPI, error) {
	klog.V(4).Infof("docker creating api client")
	c, err := client.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("error building docker client: %v", err)
	}

	if c == nil {
		return nil, fmt.Errorf("error building docker client, client returned is nil")
	}

	// Test the client
	ctx := context.Background()
	_, err = c.Info(ctx)
	if err != nil {
		// TODO check if /var/run/docker.sock exists and create a connection using that
		klog.Errorf("Unable to create docker client please set DOCKER_HOST to unix socket or tcp socket")
		klog.Errorf("Standard DOCKER_HOST values can be %q and defaults to %q", "unix:///var/run/docker.sock", client.DefaultDockerHost)
		return nil, fmt.Errorf("error building docker client, unable to make info call: %v", err)
	}

	return &dockerAPI{
		client: c,
	}, nil
}

// findImage does a `docker images` via the API, and finds the specified image
func (d *dockerAPI) findImage(name string) (*types.Image, error) {
	klog.V(4).Infof("docker query for image %q", name)
	options := types.ImageListOptions{
		MatchName: name,
	}
	ctx := context.Background()
	images, err := d.client.ImageList(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("error listing images: %v", err)
	}
	for i := range images {
		for _, repoTag := range images[i].RepoTags {
			if repoTag == name {
				return &images[i], nil
			}
		}
	}
	return nil, nil
}

// tagImage does a `docker tag`, via the API
func (d *dockerAPI) tagImage(imageID string, ref string) error {
	klog.V(4).Infof("docker tag for image %q, tag %q", imageID, ref)

	ctx := context.Background()
	options := types.ImageTagOptions{}
	err := d.client.ImageTag(ctx, imageID, ref, options)
	if err != nil {
		return fmt.Errorf("error tagging image %q with tag %q: %v", imageID, ref, err)
	}
	return nil
}
