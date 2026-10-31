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
	"context"
	"errors"
	"fmt"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/kubetest2-kops/gce"
	"k8s.io/kops/tests/e2e/pkg/kops"
	"sigs.k8s.io/kubetest2/pkg/boskos"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

func (d *deployer) Down() error {
	err := d.down()
	if d.KeepStagingStore {
		return err
	}

	// Build() creates the S3 staging bucket before there is a cluster, so it is deleted even if
	// down() found no cluster to delete, or failed to delete it.
	return errors.Join(err, d.deleteStagingStore())
}

// deleteStagingStore deletes the S3 staging bucket if the deployer created it.
func (d *deployer) deleteStagingStore() error {
	if d.CloudProvider == "aws" && d.createStagingStore {
		return d.aws.DeleteS3BucketAndObjects(context.Background(), d.stagingStore())
	}
	return nil
}

// down deletes the cluster, and the buckets that Up() created for it. Up() calls it to delete a
// cluster leaked by an earlier run. Unlike Down(), it keeps the S3 staging bucket, which Build()
// has already filled for the new cluster.
func (d *deployer) down() error {
	if err := d.init(); err != nil {
		return err
	}

	// There is no point running the rest of this function if the cluster doesn't exist.
	// The lookup needs d.env() like every other kops call: the deployer may compute the
	// state store itself, and the job does not always export KOPS_STATE_STORE. Without it
	// the lookup fails, and a cluster leaked by an earlier run is never deleted.
	cluster, _ := kops.GetCluster(d.KopsBinaryPath, d.ClusterName, d.env(), false)
	if cluster == nil {
		return nil
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
	args = append(args,
		"--interval=60s",
		"--wait=60m",
	)
	klog.Info(strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)

	exec.InheritOutput(cmd)
	if err := cmd.Run(); err != nil {
		return err
	}

	switch d.CloudProvider {
	case "aws":
		ctx := context.Background()
		if d.createStateStore {
			if err := d.aws.DeleteS3Bucket(ctx, d.stateStore()); err != nil {
				return err
			}
		}
		if d.createDiscoveryStore {
			if err := d.aws.DeleteS3Bucket(ctx, d.discoveryStore()); err != nil {
				return err
			}
		}
	case "gce":
		if d.createStateStore {
			gce.DeleteGCSBucket(d.stateStore(), d.GCPProject)
			gce.DeleteGCSBucket(d.stagingStore(), d.GCPProject)
		}
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
