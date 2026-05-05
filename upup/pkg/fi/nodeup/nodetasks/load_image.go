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

package nodetasks

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/backoff"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"k8s.io/kops/util/pkg/hashing"
)

// LoadImageTask is responsible for downloading a docker image
type LoadImageTask struct {
	Name    string
	Sources []string
	Hash    string
	Runtime string
}

var (
	_ fi.NodeupTask            = &LoadImageTask{}
	_ fi.NodeupHasDependencies = &LoadImageTask{}
)

func (t *LoadImageTask) GetDependencies(tasks map[string]fi.NodeupTask) []fi.NodeupTask {
	// LoadImageTask depends on the docker service to ensure we
	// sideload images after docker is completely updated and
	// configured.
	var deps []fi.NodeupTask
	for _, v := range tasks {
		if svc, ok := v.(*Service); ok && svc.Name == containerdService {
			deps = append(deps, v)
		}
		if svc, ok := v.(*Service); ok && svc.Name == dockerService {
			deps = append(deps, v)
		}
	}
	return deps
}

var _ fi.HasName = (*LoadImageTask)(nil)

func (t *LoadImageTask) GetName() *string {
	if t.Name == "" {
		return nil
	}
	return &t.Name
}

func (t *LoadImageTask) String() string {
	return fmt.Sprintf("LoadImageTask: %v", t.Sources)
}

func (e *LoadImageTask) Find(c *fi.NodeupContext) (*LoadImageTask, error) {
	klog.Warningf("LoadImageTask checking if image present not yet implemented")
	return nil, nil
}

func (e *LoadImageTask) Run(c *fi.NodeupContext) error {
	return fi.NodeupDefaultDeltaRunMethod(e, c)
}

func (_ *LoadImageTask) CheckChanges(a, e, changes *LoadImageTask) error {
	return nil
}

func (_ *LoadImageTask) RenderLocal(_ *local.LocalTarget, a, e, changes *LoadImageTask) error {
	hash, err := hashing.FromString(e.Hash)
	if err != nil {
		return err
	}

	urls := e.Sources
	if len(urls) == 0 {
		return fmt.Errorf("no sources specified: %v", err)
	}

	primaryURL := urls[0]
	for _, url := range urls {
		err = importContainerImage(url, hash)
		if err != nil {
			klog.Warningf("error importing image from url %q: %v", url, err)
			continue
		} else {
			break
		}
	}
	if err != nil {
		// Hack to try to avoid failed downloads causing massive bandwidth bills
		backoff.DoGlobalBackoff(fmt.Errorf("failed to import image %s: %v", primaryURL, err))
		return err
	}

	return nil
}

func importContainerImage(url string, expectedHash *hashing.Hash) error {
	responseBody, err := fi.OpenURL(url)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	imageReader, verifyHash, imageCloser, err := imageImportReader(responseBody, expectedHash)
	if err != nil {
		return err
	}

	args := containerImageImportArgs()
	human := strings.Join(args, " ")

	klog.Infof("running command %s", human)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = imageReader
	output, err := cmd.CombinedOutput()
	if imageCloser != nil {
		if closeErr := imageCloser.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	if err != nil {
		return fmt.Errorf("error loading docker image with '%s': %v: %s", human, err, string(output))
	}
	if err := verifyHash(); err != nil {
		return err
	}

	return nil
}

func containerImageImportArgs() []string {
	return []string{"ctr", "--namespace", "k8s.io", "images", "import", "--no-unpack", "-"}
}

func imageImportReader(r io.Reader, expectedHash *hashing.Hash) (io.Reader, func() error, io.Closer, error) {
	algorithm := hashing.HashAlgorithmSHA256
	if expectedHash != nil {
		algorithm = expectedHash.Algorithm
	}

	hasher := algorithm.NewHasher()
	hashedReader := io.TeeReader(r, hasher)
	imageReader, closer, err := maybeGzipReader(hashedReader)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error reading container image stream: %v", err)
	}

	verifyHash := func() error {
		actualHash := &hashing.Hash{
			Algorithm: algorithm,
			HashValue: hasher.Sum(nil),
		}
		if expectedHash != nil && !actualHash.Equal(expectedHash) {
			return fmt.Errorf("downloaded container image but hash did not match expected %q", expectedHash)
		}
		return nil
	}
	return imageReader, verifyHash, closer, nil
}
