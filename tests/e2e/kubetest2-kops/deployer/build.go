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

package deployer

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/kubetest2-kops/gce"
	"k8s.io/kops/tests/e2e/pkg/util"
	"k8s.io/kops/tests/e2e/pkg/version"
	"sigs.k8s.io/kubetest2/pkg/build"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

const (
	defaultJobName = "pull-kops-e2e-kubernetes-aws"
	defaultGCSPath = "gs://k8s-staging-kops/pulls/%v/pull-%v"
)

func (d *deployer) Build() error {
	if err := d.init(); err != nil {
		return err
	}
	results, err := d.BuildOptions.Build()
	if err != nil {
		return err
	}
	if results.KopsBaseURL != "" {
		klog.Infof("setting kops base url to %q from build results", results.KopsBaseURL)
		d.KopsBaseURL = results.KopsBaseURL
	}

	if results.KubernetesBaseURL != "" {
		klog.Infof("setting kubernetes base url to %q from build results", results.KubernetesBaseURL)
		v, err := version.ParseKubernetesVersion(results.KubernetesBaseURL)
		if err != nil {
			return err
		}
		d.KubernetesVersion = v
	}

	if d.BuildOptions.BuildKubernetes {
		build.StoreCommonBinaries(d.BuildOptions.KubeRoot, d.commonOptions.RunDir())
	}
	// Copy the kops binary into the test's RunDir to be included in the tester's PATH
	if d.KopsBinaryPath != "" && !d.BuildOptions.BuildKubernetes {
		return util.Copy(d.KopsBinaryPath, path.Join(d.commonOptions.RunDir(), "kops"))
	} else {
		return nil
	}
}

func (d *deployer) verifyBuildFlags() error {
	if d.BuildOptions.TargetBuildArch != "" {
		if !strings.HasPrefix(d.BuildOptions.TargetBuildArch, "linux/") {
			return errors.New("--target-build-arch supports linux/amd64 and linux/arm64 only")
		} else if d.BuildOptions.BuildKubernetes {
			d.BuildOptions.TargetBuildArch = "linux/amd64"
		}
	}

	if d.KopsBinaryPath != "" {
		if goPath := os.Getenv("GOPATH"); goPath != "" {
			d.KopsRoot = path.Join(goPath, "src", "k8s.io", "kops")
		} else {
			return errors.New("required --kops-root when building from source")
		}
		fi, err := os.Stat(d.KopsRoot)
		if err != nil {
			return err
		}
		if !fi.Mode().IsDir() {
			return errors.New("--kops-root must be a directory")
		}
	}
	if d.BuildOptions.BuildKubernetes {
		var KubeRoot string
		if goPath := os.Getenv("GOPATH"); goPath != "" {
			KubeRoot = path.Join(goPath, "src", "k8s.io", "kubernetes")
		} else {
			return errors.New("$GOPATH is not set, please set this variable")
		}
		fi, err := os.Stat(KubeRoot)
		if err != nil {
			return err
		}
		if !fi.Mode().IsDir() {
			return errors.New("unable to find kubernetes at $GOPATH/src/k8s.io/kubernetes")
		}
		d.BuildOptions.KubeRoot = KubeRoot
	}
	if d.StageLocation != "" {
		if !strings.HasPrefix(d.StageLocation, "gs://") {
			return errors.New("stage-location must be a gs:// path")
		}
	} else if d.boskos != nil {
		d.StageLocation = d.stagingStore()
		klog.Infof("creating staging bucket %s to hold kops/kubernetes build artifacts", d.StageLocation)
		if err := gce.EnsureGCSBucket(d.StageLocation, d.GCPProject, true); err != nil {
			return err
		}
	} else {
		stageLocation, err := defaultStageLocation(d.KopsRoot)
		if err != nil {
			return err
		}
		d.StageLocation = stageLocation
	}
	if d.KopsBaseURL == "" && os.Getenv("KOPS_BASE_URL") == "" {
		d.KopsBaseURL = strings.Replace(d.StageLocation, "gs://", "https://storage.googleapis.com/", 1)
	}

	if d.KopsVersionMarker != "" && !d.BuildOptions.BuildKubernetes {
		return errors.New("cannot use --kops-version-marker with --build")
	}

	d.BuildOptions.KopsRoot = d.KopsRoot
	d.BuildOptions.StageLocation = d.StageLocation
	return nil
}

func defaultStageLocation(kopsRoot string) (string, error) {
	jobName := os.Getenv("JOB_NAME")
	if jobName == "" {
		jobName = defaultJobName
	}

	sha := os.Getenv("PULL_PULL_SHA")
	if sha == "" {
		cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
		cmd.SetDir(kopsRoot)
		output, err := exec.CombinedOutputLines(cmd)
		if err != nil {
			return "", err
		} else if len(output) != 1 {
			return "", fmt.Errorf("unexpected output from git describe: %v", output)
		}
		sha = strings.TrimSpace(output[0])
	}
	return fmt.Sprintf(defaultGCSPath, jobName, sha), nil
}
