/*
Copyright 2020 The Kubernetes Authors.

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

package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_UnparseableVersion(t *testing.T) {
	addons := Addons{
		TypeMeta: v1.TypeMeta{
			Kind: "Addons",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
		},
		Spec: AddonsSpec{
			Addons: []*AddonSpec{
				{
					Name:    s("testaddon"),
					Version: s("1.0-kops"),
				},
			},
		},
	}

	err := addons.Verify()
	assert.EqualError(t, err, "addon \"testaddon\" has unparseable version \"1.0-kops\": Short version cannot contain PreRelease/Build meta data", "detected invalid version")
}

func s(v string) *string {
	return &v
}
