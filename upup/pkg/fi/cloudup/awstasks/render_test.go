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

package awstasks

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

type renderTest struct {
	Resource interface{}
	Expected string
}

func doRenderTests(t *testing.T, method string, cases []*renderTest) {
	outdir, err := ioutil.TempDir("/tmp", "kops-render-")
	if err != nil {
		t.Errorf("failed to create local render directory: %s", err)
		t.FailNow()
	}
	defer os.RemoveAll(outdir)

	for i, c := range cases {
		var filename string
		var target interface{}

		cloud := awsup.BuildMockAWSCloud("eu-west-2", "abc")

		switch method {
		case "RenderTerraform":
			target = terraform.NewTerraformTarget(cloud, "eu-west-2", "test", outdir, terraform.Version012, nil)
			filename = "kubernetes.tf"
		case "RenderCloudformation":
			target = cloudformation.NewCloudformationTarget(cloud, "eu-west-2", "test", outdir)
			filename = "kubernetes.json"
		default:
			t.Errorf("unknown render method: %s", method)
			t.FailNow()
		}

		// @step: build the inputs for the methods - hopefully these don't change between them
		var inputs []reflect.Value
		for _, x := range []interface{}{target, c.Resource, c.Resource, c.Resource} {
			inputs = append(inputs, reflect.ValueOf(x))
		}

		err := func() error {
			// @step: invoke the rendering method of the target
			resp := reflect.ValueOf(c.Resource).MethodByName(method).Call(inputs)
			if err := resp[0].Interface(); err != nil {
				return err.(error)
			}

			// @step: invoke the target finish up
			in := []reflect.Value{reflect.ValueOf(make(map[string]fi.Task))}
			resp = reflect.ValueOf(target).MethodByName("Finish").Call(in)
			if err := resp[0].Interface(); err != nil {
				return err.(error)
			}

			// @step: check the render is as expected
			if c.Expected != "" {
				content, err := ioutil.ReadFile(path.Join(outdir, filename))
				if err != nil {
					return err
				}
				if c.Expected != string(content) {
					diffString := diff.FormatDiff(c.Expected, string(content))
					t.Logf("diff:\n%s\n", diffString)
					t.Errorf("case %d, expected: %s\n,got: %s\n", i, c.Expected, string(content))
					//assert.Equal(t, "", string(content))
				}
			}

			return nil
		}()
		if err != nil {
			t.Errorf("case %d, did not expect an error: %s", i, err)
		}
	}
}
