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

package assets

import (
	"fmt"

	"k8s.io/klog/v2"
)

// CopyImage copies a docker image from a source registry, to a target registry,
// typically used for highly secure clusters.
type CopyImage struct {
	Name        string
	SourceImage string
	TargetImage string
}

func (e *CopyImage) Run() error {
	api, err := newDockerAPI()
	if err != nil {
		return err
	}

	cli, err := newDockerCLI()
	if err != nil {
		return err

	}

	source := e.SourceImage
	target := e.TargetImage

	klog.Infof("copying docker image from %q to %q", source, target)

	err = cli.pullImage(source)
	if err != nil {
		return fmt.Errorf("error pulling image %q: %v", source, err)
	}
	sourceImage, err := api.findImage(source)
	if err != nil {
		return fmt.Errorf("error finding image %q: %v", source, err)
	}
	if sourceImage == nil {
		return fmt.Errorf("source image %q not found", source)
	}

	err = api.tagImage(sourceImage.ID, target)
	if err != nil {
		return fmt.Errorf("error tagging image %q: %v", source, err)
	}

	err = cli.pushImage(target)
	if err != nil {
		return fmt.Errorf("error pushing image %q: %v", target, err)
	}

	return nil
}
