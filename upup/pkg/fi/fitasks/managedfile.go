package fitasks

import (
	"fmt"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/pkg/util/validation/field"
	"os"
)

//go:generate fitask -type=ManagedFile
type ManagedFile struct {
	Name     *string
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

	err = c.ClusterConfigBase.Join(location).WriteFile(data)
	if err != nil {
		return fmt.Errorf("error creating ManagedFile %q: %v", location, err)
	}

	return nil
}
