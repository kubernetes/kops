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

package fi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"k8s.io/kops/util/pkg/vfs"
)

type Resource interface {
	Open() (io.Reader, error)
}

type TemplateResource interface {
	Resource
	Curry(args []string) TemplateResource
}

func ResourcesMatch(a, b Resource) (bool, error) {
	aReader, err := a.Open()
	if err != nil {
		return false, err
	}
	defer SafeClose(aReader)

	bReader, err := b.Open()
	if err != nil {
		return false, err
	}
	defer SafeClose(bReader)

	const size = 8192
	aData := make([]byte, size)
	bData := make([]byte, size)

	for {
		aN, aErr := io.ReadFull(aReader, aData)
		if aErr != nil && aErr != io.EOF && aErr != io.ErrUnexpectedEOF {
			return false, aErr
		}

		bN, bErr := io.ReadFull(bReader, bData)
		if bErr != nil && bErr != io.EOF && bErr != io.ErrUnexpectedEOF {
			return false, bErr
		}

		if aErr == nil && bErr == nil {
			if aN != size || bN != size {
				panic("violation of io.ReadFull contract")
			}
			if !bytes.Equal(aData, bData) {
				return false, nil
			}
			continue
		}

		if aN != bN {
			return false, nil
		}

		return bytes.Equal(aData[0:aN], bData[0:bN]), nil
	}
}

func CopyResource(dest io.Writer, r Resource) (int64, error) {
	in, err := r.Open()
	if err != nil {
		if os.IsNotExist(err) {
			return 0, err
		}
		return 0, fmt.Errorf("error opening resource: %v", err)
	}
	defer SafeClose(in)

	n, err := io.Copy(dest, in)
	if err != nil {
		return n, fmt.Errorf("error copying resource: %v", err)
	}
	return n, nil
}

func ResourceAsString(r Resource) (string, error) {
	buf := new(bytes.Buffer)
	_, err := CopyResource(buf, r)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func ResourceAsBytes(r Resource) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := CopyResource(buf, r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type StringResource struct {
	s string
}

func (r *StringResource) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.s)
}

var _ Resource = &StringResource{}

func NewStringResource(s string) *StringResource {
	return &StringResource{s: s}
}

func (s *StringResource) Open() (io.Reader, error) {
	r := bytes.NewReader([]byte(s.s))
	return r, nil
}

type BytesResource struct {
	data []byte
}

// MarshalJSON is a custom marshaller so this will be printed as a string (instead of nothing)
// This is used in tests to verify the expected output.
func (b *BytesResource) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(b.data))
}

var _ Resource = &BytesResource{}

func NewBytesResource(data []byte) *BytesResource {
	return &BytesResource{data: data}
}

func (r *BytesResource) Open() (io.Reader, error) {
	reader := bytes.NewReader([]byte(r.data))
	return reader, nil
}

type FileResource struct {
	Path string
}

var _ Resource = &FileResource{}

func NewFileResource(path string) *FileResource {
	return &FileResource{Path: path}
}

func (r *FileResource) Open() (io.Reader, error) {
	in, err := os.Open(r.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("error opening file %q: %v", r.Path, err)
	}
	return in, err
}

type VFSResource struct {
	Path vfs.Path
}

var _ Resource = &VFSResource{}

func NewVFSResource(path vfs.Path) *VFSResource {
	return &VFSResource{Path: path}
}

func (r *VFSResource) Open() (io.Reader, error) {
	data, err := r.Path.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("error opening file %q: %v", r.Path, err)
	}
	b := bytes.NewBuffer(data)
	return b, err
}

// ResourceHolder is used in JSON/YAML models; it holds a resource but renders to/from a string
// After unmarshaling, the resource should be found by Name, and set on Resource
type ResourceHolder struct {
	Name     string
	Resource Resource
}

var _ Resource = &ResourceHolder{}

// Open implements the Open method of the Resource interface
func (o *ResourceHolder) Open() (io.Reader, error) {
	if o.Resource == nil {
		return nil, fmt.Errorf("ResourceHolder %q is not bound", o.Name)
	}
	return o.Resource.Open()
}

// UnmarshalJSON implements the special JSON marshaling for the resource, rendering the name
func (o *ResourceHolder) UnmarshalJSON(data []byte) error {
	var jsonName string
	err := json.Unmarshal(data, &jsonName)
	if err != nil {
		return err
	}
	o.Name = jsonName
	return nil
}

// Unwrap returns the underlying resource
func (o *ResourceHolder) Unwrap() Resource {
	return o.Resource
}

// AsString returns the value of the resource as a string
func (o *ResourceHolder) AsString() (string, error) {
	return ResourceAsString(o.Unwrap())
}

// AsString returns the value of the resource as a byte-slice
func (o *ResourceHolder) AsBytes() ([]byte, error) {
	return ResourceAsBytes(o.Unwrap())
}

// WrapResource creates a ResourceHolder for the specified resource
func WrapResource(r Resource) *ResourceHolder {
	return &ResourceHolder{
		Resource: r,
	}
}

type TaskDependentResource struct {
	Resource Resource `json:"resource,omitempty"`
	Task     Task     `json:"task,omitempty"`
}

var _ Resource = &TaskDependentResource{}
var _ HasDependencies = &TaskDependentResource{}

func (r *TaskDependentResource) Open() (io.Reader, error) {
	if r.Resource == nil {
		return nil, fmt.Errorf("resource opened before it is ready")
	}
	return r.Resource.Open()
}

func (r *TaskDependentResource) GetDependencies(tasks map[string]Task) []Task {
	return []Task{r.Task}
}
