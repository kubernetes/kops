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
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/storage/v1"
	"k8s.io/kops/cloudmock/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/util/pkg/vfs"
)

var credsFile = "./mock_gcp_credentials.json"

func TestGSRenderTerraform(t *testing.T) {
	creds, err := filepath.Abs(credsFile)
	if err != nil {
		t.Fatalf("failed to prepare mock gcp credentials: %v", err)
	}
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", creds)

	content := "hello world"
	grid := []struct {
		expectedPath       string
		gsPath             string
		gsObject           string
		serviceAcct        string
		expectedObjectJSON string
		expectedACLJSON    string
	}{
		{
			gsPath:      "gs://foo/bar",
			gsObject:    "bar",
			serviceAcct: "foo-123@project.iam.gserviceaccount.com",
			expectedObjectJSON: `
			{
				"bucket": "foo",
				"name": "bar",
				"provider": "${google.files}",
				"source": "\"${path.module}/data/google_storage_bucket_object_bar_content\""
			}
			`,
			expectedACLJSON: `
			{
				"bucket": "foo",
				"object": "${google_storage_bucket_object.bar.output_name}",
				"provider": "${google.files}",
				"role_entity": [
					"READER:user-foo-123@project.iam.gserviceaccount.com"
				]
			}
			`,
		},
	}
	for _, tc := range grid {

		t.Run(tc.gsPath, func(t *testing.T) {
			cloud := gce.InstallMockGCECloud("us-central1", "project")
			path, err := vfs.Context.BuildVfsPath(tc.gsPath)
			if err != nil {
				t.Fatalf("error building VFS path: %v", err)
			}

			vfsProvider, err := path.(vfs.TerraformPath).TerraformProvider()
			if err != nil {
				t.Fatalf("error building VFS Terraform provider: %v", err)
			}
			target := terraform.NewTerraformTarget(cloud, "", vfsProvider, "/dev/null", nil)

			acl := &vfs.GSAcl{
				Acl: []*storage.ObjectAccessControl{
					{
						Entity: fmt.Sprintf("user-%v", tc.serviceAcct),
						Role:   "READER",
					},
				},
			}
			err = path.(*vfs.GSPath).RenderTerraform(
				&target.TerraformWriter, tc.gsObject, strings.NewReader(content), acl,
			)
			if err != nil {
				t.Fatalf("error rendering terraform %v", err)
			}
			res, err := target.GetResourcesByType()
			if err != nil {
				t.Fatalf("error fetching terraform resources: %v", err)
			}
			if objs := res["google_storage_bucket_object"]; objs == nil {
				t.Fatalf("google_storage_bucket_object resources not found: %v", res)
			}
			if obj := res["google_storage_bucket_object"][tc.gsObject]; obj == nil {
				t.Fatalf("google_storage_bucket_object object not found: %v", res["google_storage_bucket_object"])
			}
			obj, err := json.Marshal(res["google_storage_bucket_object"][tc.gsObject])
			if err != nil {
				t.Fatalf("error marshaling gs object: %v", err)
			}
			if !assert.JSONEq(t, tc.expectedObjectJSON, string(obj), "JSON representation of terraform resource did not match") {
				t.FailNow()
			}
			if objs := target.TerraformWriter.Files[fmt.Sprintf("data/google_storage_bucket_object_%v_content", tc.gsObject)]; objs == nil {
				t.Fatalf("google_storage_bucket_object content file not found: %v", target.TerraformWriter.Files)
			}
			actualContent := string(target.TerraformWriter.Files[fmt.Sprintf("data/google_storage_bucket_object_%v_content", tc.gsObject)])
			if !assert.Equal(t, content, actualContent, "google_storage_bucket_object content did not match") {
				t.FailNow()
			}

			if objs := res["google_storage_object_access_control"]; objs == nil {
				t.Fatalf("google_storage_object_access_control resources not found: %v", res)
			}
			if obj := res["google_storage_object_access_control"][tc.gsObject]; obj == nil {
				t.Fatalf("google_storage_object_access_control object not found: %v", res["google_storage_object_access_control"])
			}
			actualACL, err := json.Marshal(res["google_storage_object_access_control"][tc.gsObject])
			if err != nil {
				t.Fatalf("error marshaling gs ACL: %v", err)
			}
			if !assert.JSONEq(t, tc.expectedACLJSON, string(actualACL), "JSON representation of terraform resource did not match") {
				t.FailNow()
			}

		})
	}
}
