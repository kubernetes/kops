package kubernetes

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/pkg/client/simple"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/kutil"
	"k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3"
	"fmt"
)

type KubernetesTarget struct {
	//kubectlContext string
	//keystore *k8sapi.KubernetesKeystore
	KubernetesClient release_1_3.Interface
	cluster          *kopsapi.Cluster
}

func NewKubernetesTarget(clientset simple.Clientset, keystore fi.Keystore, cluster *kopsapi.Cluster) (*KubernetesTarget, error) {
	b := &kutil.CreateKubecfg{
		ContextName: cluster.Name,
		KeyStore: keystore,
		SecretStore: nil,
		KubeMasterIP: cluster.Spec.MasterPublicName,
	}

	kubeconfig, err := b.ExtractKubeconfig()
	if err != nil {
		return nil, fmt.Errorf("error building credentials for cluster %q: %v", cluster.Name, err)
	}

	clientConfig, err := kubeconfig.BuildRestConfig()
	if err != nil {
		return nil, fmt.Errorf("error building configuration for cluster %q: %v", cluster.Name, err)
	}

	k8sClient, err := release_1_3.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot build k8s client: %v", err)
	}

	target := &KubernetesTarget{
		cluster: cluster,
		KubernetesClient: k8sClient,
	}
	return target, nil

}

var _ fi.Target = &KubernetesTarget{}

func (t *KubernetesTarget) Finish(taskMap map[string]fi.Task) error {
	return nil
}

func (t *KubernetesTarget) Apply(manifest []byte) error {
	context := t.cluster.Name

	// Would be nice if we could use RunApply from kubectl's code directly...
	// ... but that seems really hard

	kubectl := &kutil.Kubectl{}
	err := kubectl.Apply(context, manifest)
	if err != nil {
		return err
	}


	return nil
}