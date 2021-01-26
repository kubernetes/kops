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
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/pkg/kops"
)

func (d *deployer) init() error {
	var err error
	d.doInit.Do(func() { err = d.initialize() })
	return err
}

// initialize should only be called by init(), behind a sync.Once
func (d *deployer) initialize() error {
	if d.commonOptions.ShouldBuild() {
		if err := d.verifyBuildFlags(); err != nil {
			return fmt.Errorf("init failed to check build flags: %v", err)
		}
	}
	if d.commonOptions.ShouldUp() || d.commonOptions.ShouldDown() {
		if err := d.verifyKopsFlags(); err != nil {
			return fmt.Errorf("init failed to check kops flags: %v", err)
		}
	}
	if d.commonOptions.ShouldUp() {
		if err := d.verifyUpFlags(); err != nil {
			return fmt.Errorf("init failed to check up flags: %v", err)
		}
	}
	if d.KopsVersionMarker != "" {
		binaryPath, baseURL, err := kops.DownloadKops(d.KopsVersionMarker)
		if err != nil {
			return fmt.Errorf("init failed to download kops from url: %v", err)
		}
		d.KopsBinaryPath = binaryPath
		d.KopsBaseURL = baseURL
	}
	// These environment variables are defined by the "preset-aws-ssh" prow preset
	// https://github.com/kubernetes/test-infra/blob/3d3b325c98b739b526ba5d93ce21c90a05e1f46d/config/prow/config.yaml#L653-L670
	if d.SSHPrivateKeyPath == "" {
		d.SSHPrivateKeyPath = os.Getenv("AWS_SSH_PRIVATE_KEY_FILE")
	}
	if d.SSHPublicKeyPath == "" {
		d.SSHPublicKeyPath = os.Getenv("AWS_SSH_PUBLIC_KEY_FILE")
	}
	return nil
}

// verifyKopsFlags ensures common fields are set for kops commands
func (d *deployer) verifyKopsFlags() error {
	if d.ClusterName == "" {
		name, err := defaultClusterName(d.CloudProvider)
		if err != nil {
			return err
		}
		klog.Info("Using cluster name ", d.ClusterName)
		d.ClusterName = name
	}

	if d.KopsBinaryPath == "" && d.KopsVersionMarker == "" {
		if ws := os.Getenv("WORKSPACE"); ws != "" {
			d.KopsBinaryPath = path.Join(ws, "kops")
		} else {
			return errors.New("missing required --kops-binary-path when --kops-version-marker is not used")
		}
	}

	switch d.CloudProvider {
	case "aws":
	case "gce":
	case "digitalocean":
	default:
		return errors.New("unsupported --cloud-provider value")
	}

	if d.StateStore == "" {
		d.StateStore = stateStore(d.CloudProvider)
	}

	return nil
}

// env returns a list of environment variables passed to the kops binary
func (d *deployer) env() []string {
	vars := d.Env
	vars = append(vars, []string{
		fmt.Sprintf("PATH=%v", os.Getenv("PATH")),
		fmt.Sprintf("HOME=%v", os.Getenv("HOME")),
		fmt.Sprintf("KOPS_STATE_STORE=%v", d.StateStore),
		fmt.Sprintf("KOPS_FEATURE_FLAGS=%v", d.featureFlags()),
		"KOPS_RUN_TOO_NEW_VERSION=1",
	}...)
	if d.CloudProvider == "aws" {
		// Pass through some env vars if set
		for _, k := range []string{"AWS_PROFILE", "AWS_SHARED_CREDENTIALS_FILE"} {
			v := os.Getenv(k)
			if v != "" {
				vars = append(vars, k+"="+v)
			}
		}
	}
	if d.KopsBaseURL != "" {
		vars = append(vars, fmt.Sprintf("KOPS_BASE_URL=%v", d.KopsBaseURL))
	} else if baseURL := os.Getenv("KOPS_BASE_URL"); baseURL != "" {
		vars = append(vars, fmt.Sprintf("KOPS_BASE_URL=%v", os.Getenv("KOPS_BASE_URL")))
	}
	return vars
}

// featureFlags returns the kops feature flags to set
func (d *deployer) featureFlags() string {
	ff := []string{
		"+SpecOverrideFlag",
		"+AlphaAllowGCE",
	}
	val := strings.Join(ff, ",")
	for _, env := range d.Env {
		e := strings.Split(env, "=")
		if e[0] == "KOPS_FEATURE_FLAGS" && len(e) > 1 {
			val = fmt.Sprintf("%v,", e[1])
		}
	}
	return val
}

// defaultClusterName returns a kops cluster name to use when ClusterName is not set
func defaultClusterName(cloudProvider string) (string, error) {
	jobName := os.Getenv("JOB_NAME")
	buildID := os.Getenv("BUILD_ID")
	if jobName == "" || buildID == "" {
		return "", errors.New("JOB_NAME, and BUILD_ID env vars are required when --cluster-name is not set")
	}

	buildIDHash := fmt.Sprintf("%x", md5.Sum([]byte(buildID)))
	jobHash := fmt.Sprintf("%x", md5.Sum([]byte(jobName)))

	var suffix string
	switch cloudProvider {
	case "aws":
		suffix = "test-cncf-aws.k8s.io"
	default:
		suffix = "k8s.local"
	}

	return fmt.Sprintf("e2e-%v-%v.%v", buildIDHash[:10], jobHash[:5], suffix), nil
}

// stateStore returns the kops state store to use
// defaulting to values used in prow jobs
func stateStore(cloudProvider string) string {
	ss := os.Getenv("KOPS_STATE_STORE")
	if ss == "" {
		switch cloudProvider {
		case "aws":
			ss = "s3://k8s-kops-prow"
		case "gce":
			ss = "gs://k8s-kops-gce"
		}
	}
	return ss
}

// the default is $ARTIFACTS if set, otherwise ./_artifacts
// constructed as an absolute path to help the ginkgo tester because
// for some reason it needs an absolute path to the kubeconfig
func defaultArtifactsDir() (string, error) {
	if path, set := os.LookupEnv("ARTIFACTS"); set {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("failed to convert filepath from $ARTIFACTS (%s) to absolute path: %s", path, err)
		}
		return absPath, nil
	}

	absPath, err := filepath.Abs("_artifacts")
	if err != nil {
		return "", fmt.Errorf("when constructing default artifacts dir, failed to get absolute path: %s", err)
	}
	return absPath, nil
}
