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
	"os/exec"
	"path"
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/pkg/backoff"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/hashing"
)

const dockerService = "docker.service"

// LoadImageTask is responsible for downloading a docker image
type LoadImageTask struct {
	Sources []string
	Hash    string
}

var _ fi.Task = &LoadImageTask{}
var _ fi.HasDependencies = &LoadImageTask{}

func (t *LoadImageTask) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	// LoadImageTask depends on the docker service to ensure we
	// sideload images after docker is completely updated and
	// configured.
	var deps []fi.Task
	for _, v := range tasks {
		if svc, ok := v.(*Service); ok && svc.Name == dockerService {
			deps = append(deps, v)
		}
	}
	return deps
}

func (t *LoadImageTask) String() string {
	return fmt.Sprintf("LoadImageTask: %v", t.Sources)
}

func (e *LoadImageTask) Find(c *fi.Context) (*LoadImageTask, error) {
	klog.Warningf("LoadImageTask checking if image present not yet implemented")
	return nil, nil
}

func (e *LoadImageTask) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *LoadImageTask) CheckChanges(a, e, changes *LoadImageTask) error {
	return nil
}

func (_ *LoadImageTask) RenderLocal(t *local.LocalTarget, a, e, changes *LoadImageTask) error {
	hash, err := hashing.FromString(e.Hash)
	if err != nil {
		return err
	}

	urls := e.Sources
	if len(urls) == 0 {
		return fmt.Errorf("no sources specified: %v", err)
	}

	// We assume the first url is the "main" url, and download to that _name_, wherever we get it from
	primaryURL := urls[0]
	localFile := path.Join(t.CacheDir, hash.String()+"_"+utils.SanitizeString(primaryURL))

	for _, url := range urls {
		_, err = fi.DownloadURL(url, localFile, hash)
		if err != nil {
			klog.Warningf("error downloading url %q: %v", url, err)
			continue
		} else {
			break
		}
	}
	if err != nil {
		// Hack to try to avoid failed downloads causing massive bandwidth bills
		backoff.DoGlobalBackoff(fmt.Errorf("failed to download image %s: %v", primaryURL, err))
		return err
	}

	// Load the image into docker
	args := []string{"docker", "load", "-i", localFile}
	human := strings.Join(args, " ")

	klog.Infof("running command %s", human)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error loading docker image with '%s': %v: %s", human, err, string(output))
	}

	return nil
}

func (_ *LoadImageTask) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *LoadImageTask) error {
	return fmt.Errorf("LoadImageTask::RenderCloudInit not implemented")
}
