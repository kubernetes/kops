/*
Copyright 2018 The Kubernetes Authors.

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

package awstasks

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"

	"github.com/stretchr/testify/assert"
)

type terraformTest struct {
	Resource interface{}
	Expected string
}

func doTerraformRenderTests(t *testing.T, cases []*terraformTest) {
	outdir, err := ioutil.TempDir("/tmp", "terraform-render-")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.RemoveAll(outdir)

	awsup.InstallMockAWSCloud("eu-west-2", "abc")

	for i, c := range cases {
		target := terraform.NewTerraformTarget(awsup.BuildMockAWSCloud("eu-west-2", "abc"), "eu-west-2", "test", outdir, nil)

		var inputs []reflect.Value
		for _, x := range []interface{}{target, c.Resource, c.Resource, c.Resource} {
			inputs = append(inputs, reflect.ValueOf(x))
		}
		result := reflect.ValueOf(c.Resource).MethodByName("RenderTerraform").Call(inputs)
		if err := result[0].Interface(); err != nil {
			t.Errorf("case %d, did not expect an error in render: %s", i, err)
			continue
		}

		if err := target.Finish(nil); err != nil {
			t.Errorf("case %d, did not expect an error on target finish: %s", i, err)
			continue
		}

		if c.Expected != "" {
			content, err := ioutil.ReadFile(path.Join(outdir, "kubernetes.tf"))
			if err != nil {
				t.Errorf("case %d, failed to read in rendered content: %s", i, err)
				continue
			}
			if !assert.Equal(t, c.Expected, string(content)) {
				fmt.Printf("%s\n", string(content))
			}
		}
	}
}
