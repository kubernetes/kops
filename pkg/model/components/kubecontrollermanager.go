/*
Copyright 2017 The Kubernetes Authors.

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

package components

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/upup/pkg/fi/loader"
	metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"
)

const (
	defaultAttachDetachReconcileSyncPeriod = time.Minute
)

// KubeControllerManagerOptionsBuilder adds options for the kubernetes controller manager to the model.
type KubeControllerManagerOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &KubeControllerManagerOptionsBuilder{}

// BuildOptions generates the configurations used to create kubernetes controller manager manifest
func (b *KubeControllerManagerOptionsBuilder) BuildOptions(o interface{}) error {

	clusterSpec := o.(*kops.ClusterSpec)

	if clusterSpec.KubeControllerManager == nil {
		clusterSpec.KubeControllerManager = &kops.KubeControllerManagerConfig{}
	}

	k8sv148, err := util.ParseKubernetesVersion("v1.4.8")

	if err != nil {
		return fmt.Errorf("Unable to parse kubernetesVersion %s", err)
	}

	k8sv152, err := util.ParseKubernetesVersion("v1.5.2")

	if err != nil {
		return fmt.Errorf("Unable to parse kubernetesVersion %s", err)
	}

	if err != nil {
		return fmt.Errorf("Unable to parse kubernetesVersion %s", err)
	}

	kubernetesVersion, err := KubernetesVersion(clusterSpec)
	if err != nil {
		return fmt.Errorf("Unable to parse kubernetesVersion %s", err)
	}

	// In 1.4.8+ and 1.5.2+ k8s added the capability to tune the duration upon which the volume attach detach
	// component is called.
	// See https://github.com/kubernetes/kubernetes/pull/39551
	// TLDR; set this too low, and have a few EBS Volumes, and you will spam AWS api

	// if 1.4.8+ and 1.5.2+
	if (kubernetesVersion.GTE(*k8sv148) && kubernetesVersion.Minor == 4) || kubernetesVersion.GTE(*k8sv152) {

		glog.V(4).Infof("Kubernetes version %q supports AttachDetachReconcileSyncPeriod; will configure", kubernetesVersion)
		// If not set ... or set to 0s ... which is stupid
		if clusterSpec.KubeControllerManager.AttachDetachReconcileSyncPeriod == nil ||
			clusterSpec.KubeControllerManager.AttachDetachReconcileSyncPeriod.Duration.String() == "0s" {

			glog.V(8).Infof("AttachDetachReconcileSyncPeriod is not set; will set to default %v", defaultAttachDetachReconcileSyncPeriod)
			clusterSpec.KubeControllerManager.AttachDetachReconcileSyncPeriod = &metav1.Duration{Duration: defaultAttachDetachReconcileSyncPeriod}

			// If less than 1 min and greater than 1 sec ... you get a warning
		} else if clusterSpec.KubeControllerManager.AttachDetachReconcileSyncPeriod.Duration < defaultAttachDetachReconcileSyncPeriod &&
			clusterSpec.KubeControllerManager.AttachDetachReconcileSyncPeriod.Duration > time.Second {

			glog.Infof("KubeControllerManager AttachDetachReconcileSyncPeriod is set lower than recommended: %s", defaultAttachDetachReconcileSyncPeriod)

			// If less than 1sec you get an error.  Controller is coded to not allow configuration
			// less than one second.
		} else if clusterSpec.KubeControllerManager.AttachDetachReconcileSyncPeriod.Duration < time.Second {
			return fmt.Errorf("AttachDetachReconcileSyncPeriod cannot be set to less than 1 second")
		}
	} else {

		glog.V(4).Infof("not setting AttachDetachReconcileSyncPeriod, k8s version is too low")
		clusterSpec.KubeControllerManager.AttachDetachReconcileSyncPeriod = nil
	}

	return nil
}
