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
	"strings"
)

const (
	defaultJobName = "pull-kops-e2e-kubernetes-aws"
	defaultGCSPath = "gcs://kops-ci/pulls/%v"
)

func (d *deployer) Build() error {
	if err := d.init(); err != nil {
		return err
	}
	if err := d.BuildOptions.Build(); err != nil {
		return err
	}
	return nil
}

func (d *deployer) verifyBuildFlags() error {
	if d.KopsRoot == "" {
		return errors.New("required kops-root when building from source")
	}
	if d.StageLocation != "" {
		if !strings.HasPrefix(d.StageLocation, "gs://") {
			return errors.New("stage-location must be a gs:// path")
		}
	} else {
		jobName := os.Getenv("JOB_NAME")
		if jobName == "" {
			jobName = defaultJobName
		}
		d.StageLocation = fmt.Sprintf(defaultGCSPath, jobName)
	}
	fi, err := os.Stat(d.KopsRoot)
	if err != nil {
		return err
	}
	if !fi.Mode().IsDir() {
		return errors.New("kops-root must be a directory")
	}

	d.BuildOptions.KopsRoot = d.KopsRoot
	d.BuildOptions.StageLocation = d.StageLocation
	return nil
}
