/*
Copyright 2021 The Kubernetes Authors.

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

package nodetasks

import (
	"context"
	"os"
	"strings"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
)

type AptSource struct {
	Name    string
	Keyring string
	Sources []string
}

func (e *AptSource) Find(c *fi.NodeupContext) (*AptSource, error) {
	return nil, nil
}

func (f *AptSource) GetName() *string {
	return &f.Name
}

func (f *AptSource) String() string {
	return f.Name
}

func (f *AptSource) Run(c *fi.NodeupContext) error {
	return fi.NodeupDefaultDeltaRunMethod(f, c)
}

func (*AptSource) CheckChanges(a, e, changes *AptSource) error {
	return nil
}

func (f *AptSource) RenderLocal(t *local.LocalTarget, a, e, changes *AptSource) error {
	// Not adding ctx to signature as RenderLocal seems to be part of a common interface
	ctx := context.TODO()
	if _, err := fi.DownloadURL(ctx, f.Keyring, "/etc/apt/trusted.gpg.d/"+f.Name+".asc", nil); err != nil {
		return err
	}

	debs := strings.Join(f.Sources, "\n")

	if err := os.WriteFile("/etc/apt/sources.list.d/"+f.Name+".list", []byte(debs), 0); err != nil {
		return err
	}

	return nil
}
