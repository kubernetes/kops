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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/backoff"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/hashing"
)

// errCtrImport tags failures returned by the ctr import command, so the
// caller can skip DoGlobalBackoff — that backoff is meant to throttle
// repeated downloads, not failures inside containerd.
var errCtrImport = errors.New("ctr import failed")

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

func (_ *LoadImageTask) RenderLocal(t *local.LocalTarget, a, e, changes *LoadImageTask) error {
	// Not adding ctx to signature as RenderLocal seems to be part of a common interface
	ctx := context.TODO()
	hash, err := hashing.FromString(e.Hash)
	if err != nil {
		return err
	}

	urls := e.Sources
	if len(urls) == 0 {
		return fmt.Errorf("no sources specified")
	}

	if !isContainerdReady() {
		return fi.NewTryAgainLaterError("waiting for containerd to be ready")
	}

	for _, url := range urls {
		err = importContainerImage(ctx, url, hash, t.CacheDir)
		if err == nil {
			return nil
		}
		klog.Warningf("error importing image from url %q: %v", url, err)
		if errors.Is(err, errCtrImport) {
			return err
		}
	}

	// All sources failed at download. Throttle to avoid runaway bandwidth costs.
	backoff.DoGlobalBackoff(err)
	return err
}

func isContainerdReady() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ctr", "version")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

func importContainerImage(ctx context.Context, url string, expectedHash *hashing.Hash, cacheDir string) error {
	// Phase 1: fetch (or reuse cached) verified bytes. fi.DownloadURL writes via a
	// temp-file + atomic rename and short-circuits if cacheFile already matches the hash,
	// so a re-run of nodeup that already populated the cache skips the network round-trip.
	cacheFile := filepath.Join(cacheDir, expectedHash.Hex()+"_"+utils.SanitizeString(path.Base(url)))
	if _, err := fi.DownloadURL(ctx, url, cacheFile, expectedHash); err != nil {
		return err
	}

	// Phase 2: stream verified bytes (transparently gunzipped) to ctr.
	file, err := os.Open(cacheFile)
	if err != nil {
		return fmt.Errorf("error opening verified image file: %v", err)
	}
	defer file.Close()

	reader, err := maybeGzipReader(file)
	if err != nil {
		return fmt.Errorf("error decoding image archive from %q: %v", url, err)
	}
	defer reader.Close()

	args := []string{"ctr", "--namespace", "k8s.io", "images", "import", "-"}
	human := strings.Join(args, " ")

	klog.Infof("running command %s", human)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = reader
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: error loading docker image with '%s': %v: %s", errCtrImport, human, err, string(output))
	}
	return nil
}
