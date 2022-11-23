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
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
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
		fmt.Sprintf("HOME=%v", viper.GetString("HOME")),
		fmt.Sprintf("PATH=%v", viper.GetString("PATH")),
		fmt.Sprintf("GCS_LOCATION=%v", b.StageLocation),
		fmt.Sprintf("GOPATH=%v", viper.GetString("GOPATH")),
	)
	cmd.SetDir(b.KopsRoot)
	exec.InheritOutput(cmd)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Write some meta files so that other tooling can know e.g. KOPS_BASE_URL
	metaDir := filepath.Join(b.KopsRoot, ".kubetest2")

	if err := os.MkdirAll(metaDir, 0o755); err != nil {
		return fmt.Errorf("failed to Mkdir(%q): %w", metaDir, err)
	}
	p := filepath.Join(metaDir, "kops-base-url")
	kopsBaseURL := strings.Replace(b.StageLocation, "gs://", "https://storage.googleapis.com/", 1)
	if err := os.WriteFile(p, []byte(kopsBaseURL), 0o644); err != nil {
		return fmt.Errorf("failed to WriteFile(%q): %w", p, err)
	}

	return nil
}
