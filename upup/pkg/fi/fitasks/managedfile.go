/*
Copyright 2019 The Kubernetes Authors.

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

package fitasks

import (
	"bytes"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/acls"
	"k8s.io/kops/upup/pkg/fi"
)

//go:generate fitask -type=ManagedFile
type ManagedFile struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	Location *string
	Contents *fi.ResourceHolder
}

func (e *ManagedFile) Find(c *fi.Context) (*ManagedFile, error) {
	managedFiles := c.ClusterConfigBase

	location := fi.StringValue(e.Location)
	if location == "" {
		return nil, nil
	}

	existingData, err := managedFiles.Join(location).ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	actual := &ManagedFile{
		Name:     e.Name,
		Location: e.Location,
		Contents: fi.WrapResource(fi.NewBytesResource(existingData)),
	}

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *ManagedFile) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *ManagedFile) CheckChanges(a, e, changes *ManagedFile) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	if e.Contents == nil {
		return field.Required(field.NewPath("Contents"), "")
	}
	return nil
}

func (_ *ManagedFile) Render(c *fi.Context, a, e, changes *ManagedFile) error {
	location := fi.StringValue(e.Location)
	if location == "" {
		return fi.RequiredField("Location")
	}

	data, err := e.Contents.AsBytes()
	if err != nil {
		return fmt.Errorf("error reading contents of ManagedFile: %v", err)
	}

	p := c.ClusterConfigBase.Join(location)

	acl, err := acls.GetACL(p, c.Cluster)
	if err != nil {
		return err
	}

	err = p.WriteFile(bytes.NewReader(data), acl)
	if err != nil {
		return fmt.Errorf("error creating ManagedFile %q: %v", location, err)
	}

	return nil
}
