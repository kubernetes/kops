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
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/kubetest2-kops/gce"
	"k8s.io/kops/tests/e2e/pkg/target"
	"k8s.io/kops/tests/e2e/pkg/util"
	"sigs.k8s.io/kubetest2/pkg/boskos"
)

func (d *deployer) init() error {
	var err error
	d.doInit.Do(func() { err = d.initialize() })
	return err
}

// initialize should only be called by init(), behind a sync.Once
func (d *deployer) initialize() error {
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

	switch d.CloudProvider {
	case "aws":
		if d.SSHPrivateKeyPath == "" {
			d.SSHPrivateKeyPath = os.Getenv("AWS_SSH_PRIVATE_KEY_FILE")
		}
		if d.SSHPublicKeyPath == "" {
			d.SSHPublicKeyPath = os.Getenv("AWS_SSH_PUBLIC_KEY_FILE")
		}
		if d.SSHPrivateKeyPath == "" || d.SSHPublicKeyPath == "" {
			publicKeyPath, privateKeyPath, err := util.CreateSSHKeyPair(d.ClusterName)
			if err != nil {
				return err
			}
			d.SSHPublicKeyPath = publicKeyPath
			d.SSHPrivateKeyPath = privateKeyPath
		}
	case "digitalocean":
		if d.SSHPrivateKeyPath == "" {
			d.SSHPrivateKeyPath = os.Getenv("DO_SSH_PRIVATE_KEY_FILE")
		}
		if d.SSHPublicKeyPath == "" {
			d.SSHPublicKeyPath = os.Getenv("DO_SSH_PUBLIC_KEY_FILE")
		}
		d.SSHUser = "root"
	case "gce":
		if d.GCPProject == "" {
			klog.V(1).Info("No GCP project provided, acquiring from Boskos")

			boskosClient, err := boskos.NewClient(d.BoskosLocation)
			if err != nil {
				return fmt.Errorf("failed to make boskos client: %s", err)
			}
			d.boskos = boskosClient

			resource, err := boskos.Acquire(
				d.boskos,
				d.BoskosResourceType,
				d.BoskosAcquireTimeout,
				d.BoskosHeartbeatInterval,
				d.boskosHeartbeatClose,
			)
			if err != nil {
				return fmt.Errorf("init failed to get project from boskos: %s", err)
			}
			d.GCPProject = resource.Name
			klog.V(1).Infof("Got project %s from boskos", d.GCPProject)

			if d.SSHPrivateKeyPath == "" {
				d.SSHPrivateKeyPath = os.Getenv("GCE_SSH_PRIVATE_KEY_FILE")
			}
			if d.SSHPublicKeyPath == "" {
				d.SSHPublicKeyPath = os.Getenv("GCE_SSH_PUBLIC_KEY_FILE")
			}
			if d.SSHPrivateKeyPath == "" && d.SSHPublicKeyPath == "" {
				privateKey, publicKey, err := gce.SetupSSH(d.GCPProject)
				if err != nil {
					return err
				}
				d.SSHPrivateKeyPath = privateKey
				d.SSHPublicKeyPath = publicKey
			}

			d.createBucket = true
		} else if d.SSHPrivateKeyPath == "" && os.Getenv("KUBE_SSH_KEY_PATH") != "" {
			d.SSHPrivateKeyPath = os.Getenv("KUBE_SSH_KEY_PATH")
		}
	}

	klog.V(1).Infof("Using SSH keypair: [%s,%s]", d.SSHPrivateKeyPath, d.SSHPublicKeyPath)

	if d.commonOptions.ShouldBuild() {
		if err := d.verifyBuildFlags(); err != nil {
			return fmt.Errorf("init failed to check build flags: %v", err)
		}
	}

	if d.SSHUser == "" {
		d.SSHUser = os.Getenv("KUBE_SSH_USER")
	}
	klog.V(1).Infof("Using SSH user: [%s]", d.SSHUser)

	if d.TerraformVersion != "" {
		t, err := target.NewTerraform(d.TerraformVersion, d.ArtifactsDir)
		if err != nil {
			return err
		}
		d.terraform = t
	}
	if d.commonOptions.ShouldTest() {
		for _, envvar := range d.env() {
			// Set all of the env vars we use for kops in the current process
			// so that the tester inherits them when shelling out to kops
			if i := strings.Index(envvar, "="); i != -1 {
				os.Setenv(envvar[0:i], envvar[i+1:])
			} else {
				os.Setenv(envvar, "")
			}
		}
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
		d.ClusterName = name
		klog.Infof("Using cluster name: %v", d.ClusterName)
	}

	if d.KopsBinaryPath == "" && d.KopsVersionMarker == "" {
		return errors.New("missing required --kops-binary-path when --kops-version-marker is not used")
	}

	if d.ControlPlaneCount == 0 {
		d.ControlPlaneCount = 1
	}

	switch d.CloudProvider {
	case "aws":
	case "gce":
	case "digitalocean":
	default:
		return errors.New("unsupported --cloud-provider value")
	}

	return nil
}

// env returns a list of environment variables passed to the kops binary
func (d *deployer) env() []string {
	vars := d.Env
	vars = append(vars, []string{
		fmt.Sprintf("PATH=%v", os.Getenv("PATH")),
		fmt.Sprintf("HOME=%v", os.Getenv("HOME")),
		fmt.Sprintf("KOPS_STATE_STORE=%v", d.stateStore()),
		fmt.Sprintf("KOPS_FEATURE_FLAGS=%v", d.featureFlags()),
		"KOPS_RUN_TOO_NEW_VERSION=1",
	}...)

	if d.BuildOptions.TargetBuildArch != "" {
		vars = append(vars, fmt.Sprintf("KOPS_ARCH=%s", strings.Trim(d.BuildOptions.TargetBuildArch, "linux/")))
	}

	// Pass-through some env vars if set (on all clouds)
	for _, k := range []string{"KOPS_ARCH"} {
		if v := os.Getenv(k); v != "" {
			vars = append(vars, k+"="+v)
		}
	}

	if d.CloudProvider == "aws" {
		// Pass through some env vars if set
		for _, k := range []string{"AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE", "AWS_CONTAINER_CREDENTIALS_FULL_URI", "AWS_PROFILE", "AWS_SHARED_CREDENTIALS_FILE"} {
			v := os.Getenv(k)
			if v != "" {
				vars = append(vars, k+"="+v)
			}
		}
		// Recognized by the e2e framework
		// https://github.com/kubernetes/kubernetes/blob/a750d8054a6cb3167f495829ce3e77ab0ccca48e/test/e2e/framework/ssh/ssh.go#L59-L62
		vars = append(vars, fmt.Sprintf("KUBE_SSH_KEY_PATH=%v", d.SSHPrivateKeyPath))
	} else if d.CloudProvider == "digitalocean" {
		// Pass through some env vars if set
		for _, k := range []string{"DIGITALOCEAN_ACCESS_TOKEN", "S3_ACCESS_KEY_ID", "S3_SECRET_ACCESS_KEY"} {
			v := os.Getenv(k)
			if v != "" {
				vars = append(vars, k+"="+v)
			} else {
				klog.Warningf("DO env var %s is empty..", k)
			}
		}
	}
	if d.KopsBaseURL != "" {
		vars = append(vars, fmt.Sprintf("KOPS_BASE_URL=%v", d.KopsBaseURL))
	} else if baseURL := os.Getenv("KOPS_BASE_URL"); baseURL != "" {
		vars = append(vars, fmt.Sprintf("KOPS_BASE_URL=%v", os.Getenv("KOPS_BASE_URL")))
	}

	// Pass through OpenTelemetry flags
	{
		foundOTEL := false
		for _, k := range []string{
			"OTEL_EXPORTER_OTLP_TRACES_FILE", "OTEL_EXPORTER_OTLP_FILE",
			"OTEL_EXPORTER_OTLP_TRACES_DIR", "OTEL_EXPORTER_OTLP_DIR",
		} {
			v := os.Getenv(k)
			if v != "" {
				foundOTEL = true
				vars = append(vars, k+"="+v)
			}
		}

		// If no otel flags were explicitly specified, and we have artifacts, log under the artifacts directory
		if !foundOTEL {
			artifacts := d.ArtifactsDir
			if artifacts != "" {
				vars = append(vars, "OTEL_EXPORTER_OTLP_TRACES_DIR="+filepath.Join(artifacts, "otlp"))
			}
		}
	}
	return vars
}

// featureFlags returns the kops feature flags to set
func (d *deployer) featureFlags() string {
	for _, env := range d.Env {
		e := strings.Split(env, "=")
		if e[0] == "KOPS_FEATURE_FLAGS" && len(e) > 1 {
			return e[1]
		}
	}
	// if not set by the env flag, but set in the environment, use that.
	if e := os.Getenv("KOPS_FEATURE_FLAGS"); e != "" {
		return e
	}
	return ""
}

// defaultClusterName returns a kops cluster name to use when ClusterName is not set
func defaultClusterName(cloudProvider string) (string, error) {
	dnsDomain := os.Getenv("KOPS_DNS_DOMAIN")
	jobName := os.Getenv("JOB_NAME")
	jobType := os.Getenv("JOB_TYPE")
	buildID := os.Getenv("BUILD_ID")
	pullNumber := os.Getenv("PULL_NUMBER")
	if dnsDomain == "" {
		dnsDomain = "test-cncf-aws.k8s.io"
	}
	if jobName == "" || buildID == "" {
		return "", errors.New("JOB_NAME, and BUILD_ID env vars are required when --cluster-name is not set")
	}
	if jobType == "presubmit" && pullNumber == "" {
		return "", errors.New("PULL_NUMBER must be set when JOB_TYPE=presubmit and --cluster-name is not set")
	}

	var suffix string
	switch cloudProvider {
	case "aws":
		suffix = dnsDomain
	default:
		suffix = "k8s.local"
	}

	if len(jobName) > 79 { // SNS has char limit of 80
		jobName = jobName[:79]
	}
	if jobType == "presubmit" {
		jobName = fmt.Sprintf("e2e-pr%s.%s", pullNumber, jobName)
	} else {
		jobName = fmt.Sprintf("e2e-%s", jobName)
	}

	// GCP has char limit of 64
	gcpLimit := 63 - (len(suffix) + 1) // 1 for the dot
	if len(jobName) > gcpLimit && cloudProvider == "gce" {
		jobName = jobName[:gcpLimit]
	}
	return fmt.Sprintf("%v.%v", jobName, suffix), nil
}

// stateStore returns the kops state store to use
// defaulting to values used in prow jobs
func (d *deployer) stateStore() string {
	ss := os.Getenv("KOPS_STATE_STORE")
	if ss == "" {
		switch d.CloudProvider {
		case "aws":
			ss = "s3://k8s-kops-prow"
		case "gce":
			d.createBucket = true
			ss = "gs://" + gce.GCSBucketName(d.GCPProject, "state")
		case "digitalocean":
			ss = "do://e2e-kops-space"
		}
	}
	return ss
}

// discoveryStore returns the VFS path to use for public OIDC documents
func (d *deployer) discoveryStore() string {
	discovery := os.Getenv("KOPS_DISCOVERY_STORE")
	if discovery == "" {
		switch d.CloudProvider {
		case "aws":
			discovery = "s3://k8s-kops-ci-prow"
		}
	}
	return discovery
}

func (d *deployer) stagingStore() string {
	sb := os.Getenv("KOPS_STAGING_BUCKET")
	if sb == "" {
		switch d.CloudProvider {
		case "gce":
			d.createBucket = true
			sb = "gs://" + gce.GCSBucketName(d.GCPProject, "staging")
		}
	}
	return sb
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
