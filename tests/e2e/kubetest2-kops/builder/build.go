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

package builder

import (
	"fmt"
	"os"

	"sigs.k8s.io/kubetest2/pkg/exec"
)

type BuildOptions struct {
	KopsRoot      string `flag:"-"`
	StageLocation string `flag:"-"`
}

// Build will build the kops artifacts and publish them to the stage location
func (b *BuildOptions) Build() error {
	cmd := exec.Command("make", "gcs-publish-ci")
	cmd.SetEnv(
		fmt.Sprintf("HOME=%v", os.Getenv("HOME")),
		fmt.Sprintf("PATH=%v", os.Getenv("PATH")),
		fmt.Sprintf("GCS_LOCATION=%v", b.StageLocation),
	)
	cmd.SetDir(b.KopsRoot)
	exec.InheritOutput(cmd)
	return cmd.Run()
}
