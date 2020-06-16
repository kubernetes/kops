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

package nodetasks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/kops/upup/pkg/fi"
)

func TestIssueCertFileDependencies(t *testing.T) {
	context := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	issue := &IssueCert{Name: "testCert"}
	context.AddTask(issue)
	err := issue.AddFileTasks(context, "/tmp", "testCert", "testCa", nil)
	assert.NoError(t, err)
	var taskNames []string
	for name := range context.Tasks {
		taskNames = append(taskNames, name)
	}
	assert.ElementsMatch(t, []string{"IssueCert/testCert", "File//tmp", "File//tmp/testCert.crt", "File//tmp/testCert.key", "File//tmp/testCa.crt"}, taskNames)

	for _, fileName := range []string{"/tmp/testCert.crt", "/tmp/testCert.key", "/tmp/testCa.crt"} {
		task := context.Tasks["File/"+fileName]
		if !assert.NotNil(t, task) {
			continue
		}
		deps := task.(fi.HasDependencies).GetDependencies(context.Tasks)

		taskNames = nil
		for _, task := range deps {
			taskNames = append(taskNames, *task.(fi.HasName).GetName())
		}
		assert.ElementsMatch(t, []string{"testCert", "/tmp"}, taskNames)
	}
}
