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
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
	krel "k8s.io/release/pkg/build"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

type BuildOptions struct {
	KopsRoot      string `flag:"-"`
	KubeRoot      string `flag:"-"`
	StageLocation string `flag:"-"`
	// StageRegion is the AWS region of the bucket when StageLocation is an s3:// path.
	StageRegion string `flag:"-"`
	// StageEndpoint is the S3-compatible endpoint outside AWS, such as DigitalOcean Spaces, of the
	// bucket when StageLocation is an s3:// path.
	StageEndpoint   string `flag:"-"`
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
	gcsLocation := b.StageLocation
	if !strings.HasSuffix(gcsLocation, "/") {
		gcsLocation += "/"
	}

	results := &BuildResults{}
	ctx := context.Background()

	if b.BuildKubernetes {
		bucket, bucketRoot, kubeBaseURL, err := b.kubeStageLocation(ctx)
		if err != nil {
			return nil, err
		}

		if err := krel.NewInstance(&krel.Options{
			Bucket:             bucket,
			BucketRoot:         bucketRoot,
			AllowDup:           true,
			CI:                 true,
			NoUpdateLatest:     false,
			RepoRoot:           b.KubeRoot,
			KubeBuildPlatforms: b.TargetBuildArch,
		}).Build(); err != nil {
			return nil, fmt.Errorf("stage via krel push: %w", err)
		}

		results = &BuildResults{
			KubernetesBaseURL: kubeBaseURL,
		}
		return results, nil
	}

	env := []string{
		fmt.Sprintf("HOME=%v", os.Getenv("HOME")),
		fmt.Sprintf("PATH=%v", os.Getenv("PATH")),
		fmt.Sprintf("GOPATH=%v", os.Getenv("GOPATH")),
	}
	// We need to "forward" some variables, in particular the variables like "CI" that change the version we use
	for _, k := range []string{"CI"} {
		v := os.Getenv(k)
		if v != "" {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}
	if strings.HasPrefix(b.StageLocation, "s3://") {
		store, err := b.s3Store()
		if err != nil {
			return nil, err
		}
		if err := b.publishKops(ctx, env, store); err != nil {
			return nil, err
		}
	} else {
		env = append(env, fmt.Sprintf("GCS_LOCATION=%v", gcsLocation))
		if err := b.runKopsMake("gcs-publish-ci", env); err != nil {
			return nil, err
		}
	}

	// Get the full path (including subdirectory) that we uploaded to
	// It is written by gcs-publish-ci (or publishKops) to .build/upload/latest-ci.txt
	latestPath := filepath.Join(b.KopsRoot, ".build", "upload", "latest-ci.txt")
	kopsBaseURL, err := os.ReadFile(latestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", latestPath, err)
	}
	u, err := url.Parse(strings.TrimSpace(string(kopsBaseURL)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse url %q from file %q: %w", string(kopsBaseURL), latestPath, err)
	}
	u.Path = strings.ReplaceAll(u.Path, "//", "/")
	results = &BuildResults{
		KopsBaseURL: u.String(),
	}

	// Write some meta files so that other tooling can know e.g. KOPS_BASE_URL
	metaDir := filepath.Join(b.KopsRoot, ".kubetest2")
	if err := os.MkdirAll(metaDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to Mkdir(%q): %w", metaDir, err)
	}
	p := filepath.Join(metaDir, "kops-base-url")
	if err := os.WriteFile(p, []byte(results.KopsBaseURL), 0o600); err != nil {
		return nil, fmt.Errorf("failed to WriteFile(%q): %w", p, err)
	}
	klog.Infof("wrote file %q with %q", p, results.KopsBaseURL)

	return results, nil
}

// kubeStageLocation returns the krel bucket and bucket root to stage Kubernetes to, under
// <stage location>/kubernetes, and the URL of its version marker.
func (b *BuildOptions) kubeStageLocation(ctx context.Context) (bucket, bucketRoot, markerURL string, err error) {
	// krel can push to S3, but neither to another endpoint nor with public-read ACLs.
	if b.StageEndpoint != "" {
		return "", "", "", fmt.Errorf("building Kubernetes is not supported with a stage location at %s", b.StageEndpoint)
	}
	if strings.HasPrefix(b.StageLocation, "s3://") {
		return b.s3KubeStageLocation(ctx)
	}

	// krel takes GCS buckets by name.
	bucket, prefix, _ := strings.Cut(strings.TrimPrefix(b.StageLocation, "gs://"), "/")
	if bucket == "" {
		return "", "", "", fmt.Errorf("stage location %q does not contain a bucket name", b.StageLocation)
	}
	bucketRoot = path.Join(prefix, "kubernetes")
	return bucket, bucketRoot, "https://storage.googleapis.com/" + path.Join(bucket, bucketRoot, "latest.txt"), nil
}

// runKopsMake runs the make target in the kops repository with only the given env vars.
func (b *BuildOptions) runKopsMake(target string, env []string) error {
	cmd := exec.Command("make", target)
	cmd.SetEnv(env...)
	cmd.SetDir(b.KopsRoot)
	exec.InheritOutput(cmd)
	klog.Infof("Executing %q (in %v) with env %v", "make "+target, b.KopsRoot, env)
	return cmd.Run()
}
