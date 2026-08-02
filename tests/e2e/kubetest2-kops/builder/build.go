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
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"k8s.io/klog/v2"
	krel "k8s.io/release/pkg/build"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

type BuildOptions struct {
	KopsRoot        string `flag:"-"`
	KubeRoot        string `flag:"-"`
	StageLocation   string `flag:"-"`
	TargetBuildArch string `flag:"~target-build-arch" desc:"CPU architecture to test against"`
	BuildKubernetes bool   `flag:"~build-kubernetes" desc:"Set this flag to true to build kubernetes"`
}

// BuildResults describes the outcome of a successful build.
type BuildResults struct {
	KopsBaseURL       string
	KubernetesBaseURL string
}

// Build will build the kops artifacts and publish them to the stage location
func (b *BuildOptions) Build() (*BuildResults, error) {
	// We expect to upload to a subdirectory with a version identifier
	stageLocation := b.StageLocation
	if !strings.HasSuffix(stageLocation, "/") {
		stageLocation += "/"
	}

	results := &BuildResults{}

	if b.BuildKubernetes {
		// Build k/k
		re := regexp.MustCompile(`^gs://([\w-]+)/(devel|ci)(/.*)?`)

		// StageLocation is often just the root of the bucket. the leading slash has been stripped
		kubeStageLocation := b.StageLocation + "/ci/kubernetes"
		mat := re.FindStringSubmatch(kubeStageLocation)
		if mat == nil || len(mat) < 4 {
			return nil, fmt.Errorf("invalid stage location: %v. Use gs://<bucket>/<ci|devel>/<optional-suffix>", kubeStageLocation)
		}

		if err := krel.NewInstance(&krel.Options{
			Bucket:             mat[1],
			GCSRoot:            "kubernetes",
			AllowDup:           true,
			CI:                 true,
			NoUpdateLatest:     false,
			RepoRoot:           b.KubeRoot,
			KubeBuildPlatforms: b.TargetBuildArch,
		}).Build(); err != nil {
			return nil, fmt.Errorf("stage via krel push: %w", err)
		}
		kubeBaseURL := "https://storage.googleapis.com/" + mat[1] + "/kubernetes/latest.txt"

		results = &BuildResults{
			KubernetesBaseURL: kubeBaseURL,
		}
		return results, nil
	}

	target, publishEnv, err := publishConfig(stageLocation)
	if err != nil {
		return nil, err
	}

	args := []string{"make", target}
	cmd := exec.Command(args[0], args[1:]...)
	var env []string
	if strings.HasPrefix(stageLocation, "s3://") {
		// The AWS credential chain may use profiles, web identity, or container credentials.
		env = os.Environ()
	} else {
		env = []string{
			"HOME=" + os.Getenv("HOME"),
			"PATH=" + os.Getenv("PATH"),
			"GOPATH=" + os.Getenv("GOPATH"),
		}
		if ci := os.Getenv("CI"); ci != "" {
			env = append(env, "CI="+ci)
		}
	}
	cmd.SetEnv(append(env, publishEnv...)...)
	cmd.SetDir(b.KopsRoot)
	exec.InheritOutput(cmd)
	klog.Infof("Executing %q in %v with stage location %s", strings.Join(args, " "), b.KopsRoot, stageLocation)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// The publish target records the uploaded path in .build/upload/latest-ci.txt.
	latestPath := filepath.Join(b.KopsRoot, ".build", "upload", "latest-ci.txt")
	kopsBaseURL, err := os.ReadFile(latestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", latestPath, err)
	}
	baseURL, err := publishedBaseURL(string(kopsBaseURL))
	if err != nil {
		return nil, fmt.Errorf("reading URL from file %q: %w", latestPath, err)
	}
	results = &BuildResults{
		KopsBaseURL: baseURL,
	}

	// Write some meta files so that other tooling can know e.g. KOPS_BASE_URL
	metaDir := filepath.Join(b.KopsRoot, ".kubetest2")
	if err := os.MkdirAll(metaDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to Mkdir(%q): %w", metaDir, err)
	}
	p := filepath.Join(metaDir, "kops-base-url")
	if err := os.WriteFile(p, []byte(results.KopsBaseURL), 0o644); err != nil {
		return nil, fmt.Errorf("failed to WriteFile(%q): %w", p, err)
	}
	klog.Infof("wrote file %q with %q", p, results.KopsBaseURL)

	return results, nil
}

func publishConfig(stageLocation string) (string, []string, error) {
	switch {
	case strings.HasPrefix(stageLocation, "gs://"):
		return "gcs-publish-ci", []string{"GCS_LOCATION=" + stageLocation}, nil
	case strings.HasPrefix(stageLocation, "s3://"):
		return "s3-publish-ci", []string{
			"UPLOAD_DEST=" + stageLocation,
			"UPLOAD_ARGS=",
		}, nil
	default:
		return "", nil, fmt.Errorf("unsupported stage location %q", stageLocation)
	}
}

func publishedBaseURL(value string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return "", fmt.Errorf("parsing URL %q: %w", value, err)
	}
	u.Path = strings.ReplaceAll(u.Path, "//", "/")
	if u.Scheme == "s3" {
		u.Scheme = "https"
		u.Host += ".s3.amazonaws.com"
	}
	return u.String(), nil
}
