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

package bootstrap

import (
	"path/filepath"
	"testing"

	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
)

func TestBootstarapBuilder_Simple(t *testing.T) {
	t.Setenv("AWS_REGION", "us-test1")

	runInstallBuilderTest(t, "tests/simple")
}

func runInstallBuilderTest(t *testing.T, basedir string) {
	installation := Installation{
		Command: []string{"/opt/kops/bin/nodeup", "--conf=/opt/kops/conf/kube_env.yaml", "--v=8"},
	}
	tasks := make(map[string]fi.Task)
	buildContext := &fi.ModelBuilderContext{
		Tasks: tasks,
	}
	installation.Build(buildContext)

	testutils.ValidateTasks(t, filepath.Join(basedir, "tasks.yaml"), buildContext)
}
