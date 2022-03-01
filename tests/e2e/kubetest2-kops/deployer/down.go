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
	"fmt"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/kubetest2-kops/gce"
	"sigs.k8s.io/kubetest2/pkg/boskos"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

func (d *deployer) Down() error {
	if err := d.init(); err != nil {
		return err
	}
	if err := d.DumpClusterLogs(); err != nil {
		klog.Warningf("Dumping cluster logs at the start of Down() failed: %s", err)
	}

	if d.terraform != nil {
		if err := d.terraform.Destroy(); err != nil {
			return err
		}
	}

	args := []string{
		d.KopsBinaryPath, "delete", "cluster",
		"--name", d.ClusterName,
		"--yes",
	}
	klog.Info(strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)

	exec.InheritOutput(cmd)
	if err := cmd.Run(); err != nil {
		return err
	}

	if d.CloudProvider == "gce" && d.createBucket {
		gce.DeleteGCSBucket(d.stateStore(), d.GCPProject)
	}

	if d.boskos != nil {
		klog.V(2).Info("releasing boskos project")
		err := boskos.Release(
			d.boskos,
			[]string{d.GCPProject},
			d.boskosHeartbeatClose,
		)
		if err != nil {
			return fmt.Errorf("down failed to release boskos project: %s", err)
		}
	}
	return nil
}
