package components

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/loader"
	metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"
)

const (
	defaultAttachDetachReconcileSyncPeriod = time.Minute
)

// KubeControllerManagerOptionsBuilder adds options for the k-c-m to the model
type KubeControllerManagerOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &KubeControllerManagerOptionsBuilder{}

// BuildOptions tests for options to be added to the model
func (b *KubeControllerManagerOptionsBuilder) BuildOptions(o interface{}) error {
	options := o.(*kops.ClusterSpec)

	if options.KubeControllerManager == nil {
		options.KubeControllerManager = &kops.KubeControllerManagerConfig{}
	}

	k8sv148, err := kops.ParseKubernetesVersion("v1.4.8")

	if err != nil {
		return fmt.Errorf("Unable to parse kubernetesVersion v1.4.8")
	}

	k8sv152, _ := kops.ParseKubernetesVersion("v1.5.2")

	if err != nil {
		return fmt.Errorf("Unable to parse kubernetesVersion v1.5.2")
	}

	kubernetesVersion, err := b.Context.KubernetesVersion()
	if err != nil {
		return fmt.Errorf("Unable to parse kubernetesVersion")
	}

	// In 1.4.8+ and 1.5.2+ k8s added the capability to tune the duration upon which the volume attach detach
	// component is called.
	// See https://github.com/kubernetes/kubernetes/pull/39551
	// TLDR; set this too low, and have a few PVC, and you will spam AWS api

	// if 1.4.8+ and 1.5.2+
	if kubernetesVersion.GTE(*k8sv148) || kubernetesVersion.GTE(*k8sv152) {

		// If not set ... or set to 0s ... which is stupid
		if options.KubeControllerManager.AttachDetachReconcileSyncPeriod == nil ||
			options.KubeControllerManager.AttachDetachReconcileSyncPeriod.Duration.String() == "0s" {

			glog.V(8).Info("k-c-m default-attach-detach-reconcile-sync-period flag is set to defatul")
			options.KubeControllerManager.AttachDetachReconcileSyncPeriod = &metav1.Duration{Duration: defaultAttachDetachReconcileSyncPeriod}

			// If less than 1 min and greater than 1 sec ... you get a warning
		} else if options.KubeControllerManager.AttachDetachReconcileSyncPeriod.Duration < defaultAttachDetachReconcileSyncPeriod &&
			options.KubeControllerManager.AttachDetachReconcileSyncPeriod.Duration > time.Second {

			glog.Info("k-c-m default-attach-detach-reconcile-sync-period flag is set lower than recommended")

			// If less than 1sec you get an error.  Controller no worky .. it goes boom.
		} else if options.KubeControllerManager.AttachDetachReconcileSyncPeriod.Duration < time.Second {
			return fmt.Errorf("Unable to set k-c-m default-attach-detach-reconcile-sync-period flag lower than 1 second")
		}
	} else {

		glog.V(8).Info("not setting k-c-m default-attach-detach-reconcile-sync-period, k8s version is too low")
		options.KubeControllerManager.AttachDetachReconcileSyncPeriod = nil
	}

	return nil
}
