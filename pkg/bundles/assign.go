package bundles

import (
	"fmt"

	"github.com/blang/semver"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
)

// AssignBundle will assign a bundle for the current cluster; upgrade will later advise on recommended upgrades to the bundle version
func AssignBundle(c *kops.Cluster) error {
	k8sVersion, err := KubernetesVersion(c)
	if err != nil {
		return err
	}

	if c.Spec.Bundle == "" {
		channel, err := kops.LoadChannel(c.Spec.Channel)
		if err != nil {
			return err
		}
		bundleSpec := channel.FindBundleVersion(k8sVersion)
		if bundleSpec == nil {
			return fmt.Errorf("cannot find current bundle in channel %s", c.Spec.Channel)
		}
		c.Spec.Bundle = bundleSpec.Bundle
	}

	return nil
}

func KubernetesVersion(c *kops.Cluster) (semver.Version, error) {
	k8sVersion, err := util.ParseKubernetesVersion(c.Spec.KubernetesVersion)
	if err != nil || k8sVersion == nil {
		return semver.Version{}, fmt.Errorf("unable to parse KubernetesVersion %q", c.Spec.KubernetesVersion)
	}
	return *k8sVersion, nil
}

// AssignComponentVersions will "lock" the component versions in the cluster; upgrade will later advise on recommended upgrades
func AssignComponentVersions(c *kops.Cluster) error {
	k8sVersion, err := KubernetesVersion(c)
	if err != nil {
		return err
	}

	for _, etcdCluster := range c.Spec.EtcdClusters {
		if etcdCluster.Provider != kops.EtcdProviderTypeManager {
			continue
		}

		if etcdCluster.Manager == nil {
			etcdCluster.Manager = &kops.EtcdManagerSpec{}
		}

		/*
			I THINK WE NEED A TOP LEVEL RELEASE, SO WE DON'T UNCONTROLLABLY UPDATE

			EVENTUALLY kubernetesVersion becomes part of the release

			We can say "an update is available", just as we do today for kubernetes releases

			I guess at a particular kops / kubernetes release we switch the kubernetes version logic to be release version logic

			stable has:

			  releases:
			  - range: ">=1.11.0"
			    recommendedVersion: 1.11.20180729


			Then release/1.11.20180729 has:


			components:
			- name: etcdmanager
			  release: etcdmanager/1.0.20180729


			Or we could combined them into a single file, with kubernetesVersion selection
		*/

		componentName := "etcdmanager"
		if etcdCluster.Manager.Bundle == "" {
			bundleName := c.Spec.Bundle
			if bundleName == "" {
				return fmt.Errorf("bundle not set; cannot assign bundle to %s", componentName)
			}

			bundle, err := LoadBundle(&c.Spec, c.Spec.Bundle)
			if err != nil {
				return fmt.Errorf("error loading bundle %s: %v", bundleName, err)
			}

			component := bundle.FindComponent(componentName, k8sVersion)
			if component == nil {
				return fmt.Errorf("cannot find %q component in bundle %s", componentName, bundleName)
			}
			etcdCluster.Manager.Bundle = component.Location()
		}

	}

	return nil
}
