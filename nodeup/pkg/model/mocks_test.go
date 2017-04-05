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

package model

import (
	"bytes"
	"io"
	"k8s.io/kops/upup/pkg/fi"
)

type mockAssetStore struct {
}

var _ fi.AssetStore = &mockAssetStore{}

func (a *mockAssetStore) Find(key string, assetPath string) (fi.Resource, error) {
	return &mockResource{Key: key}, nil
}

type mockResource struct {
	Key string
}

var _ fi.Resource = &mockResource{}

func (r *mockResource) Open() (io.Reader, error) {
	b := &bytes.Buffer{}
	b.WriteString("MOCK:" + r.Key)
	return b, nil
}
