/*
Copyright 2021 The Kubernetes Authors.

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
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/pkg/util"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

func (d *deployer) PostTest(testErr error) error {
	if testErr != nil || d.PublishVersionMarker == "" {
		return nil
	}
	if !strings.HasPrefix(d.PublishVersionMarker, "gs://") {
		return fmt.Errorf("unsupported --publish-version-marker protocol: %v", d.PublishVersionMarker)
	}
	if d.KopsVersionMarker == "" {
		return errors.New("missing required --kops-version-marker for publishing to --publish-version-marker")
	}

	var b bytes.Buffer
	if err := util.HTTPGETWithHeaders(d.KopsVersionMarker, nil, &b); err != nil {
		return err
	}
	tempSrc, err := os.CreateTemp("", "kops-version-marker")
	if err != nil {
		return err
	}
	_, err = tempSrc.WriteString(b.String())
	if err != nil {
		return err
	}

	args := []string{
		"gsutil",
		"-h", "Cache-Control:private, max-age=0, no-transform",
		"cp",
		tempSrc.Name(),
		d.PublishVersionMarker,
	}
	klog.Info(strings.Join(args, " "))

	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)

	exec.InheritOutput(cmd)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
