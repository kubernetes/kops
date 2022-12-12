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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"k8s.io/kops/util/pkg/vfs"
)

type Resource interface {
	Open(ctx context.Context) (io.Reader, error)
}

// HasIsReady is implemented by Resources that are derived (and thus may not be ready at comparison time)
type HasIsReady interface {
	IsReady() bool
}

func ResourcesMatch(ctx context.Context, a, b Resource) (bool, error) {
	aReader, err := a.Open(ctx)
	if err != nil {
		return false, err
	}
	defer SafeClose(aReader)

	bReader, err := b.Open(ctx)
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

func CopyResource(ctx context.Context, dest io.Writer, r Resource) (int64, error) {
	in, err := r.Open(ctx)
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

func ResourceAsString(ctx context.Context, r Resource) (string, error) {
	buf := new(bytes.Buffer)
	_, err := CopyResource(ctx, buf, r)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func ResourceAsBytes(ctx context.Context, r Resource) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := CopyResource(ctx, buf, r)
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

func (s *StringResource) Open(ctx context.Context) (io.Reader, error) {
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

func (r *BytesResource) Open(ctx context.Context) (io.Reader, error) {
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

func (r *FileResource) Open(ctx context.Context) (io.Reader, error) {
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

func (r *VFSResource) Open(ctx context.Context) (io.Reader, error) {
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

type TaskDependentResource struct {
	Resource Resource `json:"resource,omitempty"`
	Task     Task     `json:"task,omitempty"`
}

var (
	_ Resource        = &TaskDependentResource{}
	_ HasDependencies = &TaskDependentResource{}
	_ HasIsReady      = &TaskDependentResource{}
)

func (r *TaskDependentResource) Open(ctx context.Context) (io.Reader, error) {
	if r.Resource == nil {
		return nil, fmt.Errorf("resource opened before it is ready (task=%v)", r.Task)
	}
	return r.Resource.Open(ctx)
}

func (r *TaskDependentResource) GetDependencies(tasks map[string]Task) []Task {
	return []Task{r.Task}
}

// IsReady implements HasIsReady::IsReady
func (r *TaskDependentResource) IsReady() bool {
	return r.Resource != nil
}

// FunctionToResource converts a function to a Resource.  The result of executing the function is cached.
func FunctionToResource(fn func() ([]byte, error)) Resource {
	return &functionResource{
		fn: fn,
	}
}

type functionResource struct {
	data []byte
	fn   func() ([]byte, error)
}

func (r *functionResource) Open(ctx context.Context) (io.Reader, error) {
	b := r.data
	if b == nil {
		data, err := r.fn()
		if err != nil {
			return nil, err
		}
		r.data = data
		b = data
	}
	return bytes.NewReader(b), nil
}
