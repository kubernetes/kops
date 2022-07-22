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

package tests

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/util/pkg/vfs"
)

func TestS3RenderTerraform(t *testing.T) {
	content := "hello world"
	grid := []struct {
		expectedPath string
		s3Path       string
		s3Object     string
		expectedJSON string
	}{
		{
			s3Path:   "s3://foo/bar",
			s3Object: "bar",
			expectedJSON: `
			{
				"acl": "bucket-owner-full-control",
				"bucket": "foo",
				"content": "${file(\"${path.module}/data/aws_s3_object_bar_content\")}",
				"key": "bar",
				"provider": "${aws.files}",
				"server_side_encryption": "AES256"
			}
			`,
		},
	}
	t.Setenv("S3_ENDPOINT", "foo.s3.amazonaws.com")

	t.Setenv("KOPS_STATE_S3_ACL", "bucket-owner-full-control")
	for _, tc := range grid {

		t.Run(tc.s3Path, func(t *testing.T) {
			cloud := awsup.BuildMockAWSCloud("us-east-1", "a")
			path, err := vfs.Context.BuildVfsPath(tc.s3Path)
			if err != nil {
				t.Fatalf("error building VFS path: %v", err)
			}

			vfsProvider, err := path.(vfs.TerraformPath).TerraformProvider()
			if err != nil {
				t.Fatalf("error building VFS Terraform provider: %v", err)
			}
			target := terraform.NewTerraformTarget(cloud, "", vfsProvider, "/dev/null", nil)

			err = path.(*vfs.S3Path).RenderTerraform(
				&target.TerraformWriter, tc.s3Object, strings.NewReader(content), vfs.S3Acl{},
			)
			if err != nil {
				t.Fatalf("error rendering terraform %v", err)
			}
			res, err := target.GetResourcesByType()
			if err != nil {
				t.Fatalf("error fetching terraform resources: %v", err)
			}
			if objs := res["aws_s3_object"]; objs == nil {
				t.Fatalf("aws_s3_object resources not found: %v", res)
			}
			if obj := res["aws_s3_object"][tc.s3Object]; obj == nil {
				t.Fatalf("aws_s3_object object not found: %v", res["aws_s3_object"])
			}
			obj, err := json.Marshal(res["aws_s3_object"][tc.s3Object])
			if err != nil {
				t.Fatalf("error marshaling s3 object: %v", err)
			}
			if !assert.JSONEq(t, tc.expectedJSON, string(obj), "JSON representation of terraform resource did not match") {
				t.FailNow()
			}
			if objs := target.TerraformWriter.Files[fmt.Sprintf("data/aws_s3_object_%v_content", tc.s3Object)]; objs == nil {
				t.Fatalf("aws_s3_object content file not found: %v", target.TerraformWriter.Files)
			}
			actualContent := string(target.TerraformWriter.Files[fmt.Sprintf("data/aws_s3_object_%v_content", tc.s3Object)])
			if !assert.Equal(t, content, actualContent, "aws_s3_object content did not match") {
				t.FailNow()
			}
		})
	}
}
