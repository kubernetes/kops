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

	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
)

// CopyDockerImage copies a docker image from a source registry, to a target registry,
// typically used for highly secure clusters.
//go:generate fitask -type=CopyDockerImage
type CopyDockerImage struct {
	Name        *string
	SourceImage *string
	TargetImage *string
	Lifecycle   *fi.Lifecycle
}

var _ fi.CompareWithID = &CopyDockerImage{}

func (e *CopyDockerImage) CompareWithID() *string {
	return e.Name
}

func (e *CopyDockerImage) Find(c *fi.Context) (*CopyDockerImage, error) {
	return nil, nil

	// The problem here is that we can tag a local image with the remote tag, but there is no way to know
	// if that has actually been pushed to the remote registry without doing a docker push

	// The solution is probably to query the registries directly, but that is a little bit more code...

	// For now, we just always do the copy; it isn't _too_ slow when things have already been pushed

	//d, err := newDocker()
	//if err != nil {
	//	return nil, err
	//}
	//
	//source := fi.StringValue(e.SourceImage)
	//target := fi.StringValue(e.TargetImage)
	//
	//targetImage, err := d.findImage(target)
	//if err != nil {
	//	return nil, err
	//}
	//if targetImage == nil {
	//	klog.V(4).Infof("target image %q not found", target)
	//	return nil, nil
	//}
	//
	//// We want to verify that the target image matches
	//if err := d.pullImage(source); err != nil {
	//	return nil, err
	//}
	//
	//sourceImage, err := d.findImage(source)
	//if err != nil {
	//	return nil, err
	//}
	//if sourceImage == nil {
	//	return nil, fmt.Errorf("source image %q not found", source)
	//}
	//
	//if sourceImage.ID == targetImage.ID {
	//	actual := &CopyDockerImage{}
	//	actual.Name = e.Name
	//	actual.SourceImage = e.SourceImage
	//	actual.TargetImage = e.TargetImage
	//	klog.Infof("Found image %q = %s", target, sourceImage.ID)
	//	return actual, nil
	//}
	//
	//klog.V(2).Infof("Target image %q does not match source %q: %q vs %q",
	//	target, source,
	//	targetImage.ID, sourceImage.ID)
	//
	//return nil, nil
}

func (e *CopyDockerImage) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *CopyDockerImage) CheckChanges(a, e, changes *CopyDockerImage) error {
	if fi.StringValue(e.Name) == "" {
		return fi.RequiredField("Name")
	}
	if fi.StringValue(e.SourceImage) == "" {
		return fi.RequiredField("SourceImage")
	}
	if fi.StringValue(e.TargetImage) == "" {
		return fi.RequiredField("TargetImage")
	}
	return nil
}

func (_ *CopyDockerImage) Render(c *fi.Context, a, e, changes *CopyDockerImage) error {
	api, err := newDockerAPI()
	if err != nil {
		return err
	}

	cli, err := newDockerCLI()
	if err != nil {
		return err

	}

	source := fi.StringValue(e.SourceImage)
	target := fi.StringValue(e.TargetImage)

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
