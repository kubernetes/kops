/*
Copyright 2023 The Kubernetes Authors.

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

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	api "k8s.io/kops/clusterapi/bootstrap/kops/api/v1beta1"
	clusterv1 "k8s.io/kops/clusterapi/snapshot/cluster-api/api/v1beta1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/resources"
	"k8s.io/kops/pkg/nodemodel/wellknownassets"
	"k8s.io/kops/pkg/wellknownservices"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/mirrors"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NewKopsConfigReconciler is the constructor for a KopsConfigReconciler
func NewKopsConfigReconciler(mgr manager.Manager) error {
	r := &KopsConfigReconciler{
		client: mgr.GetClient(),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&api.KopsConfig{}).
		Complete(r)
}

// KopsConfigReconciler observes KopsConfig objects.
type KopsConfigReconciler struct {
	// client is the controller-runtime client
	client client.Client
}

// +kubebuilder:rbac:groups=,resources=nodes,verbs=get;list;watch;patch

// Reconcile is the main reconciler function that observes node changes.
func (r *KopsConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	obj := &api.KopsConfig{}
	if err := r.client.Get(ctx, req.NamespacedName, obj); err != nil {
		klog.Warningf("unable to fetch object: %v", err)
		if apierrors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	data, err := r.buildBootstrapData(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.storeBootstrapData(ctx, obj, data); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.client.Status().Update(ctx, obj); err != nil {
		return ctrl.Result{}, fmt.Errorf("error patching status: %w", err)
	}
	return ctrl.Result{}, nil
}

// storeBootstrapData creates a new secret with the data passed in as input,
// sets the reference in the configuration status and ready to true.
func (r *KopsConfigReconciler) storeBootstrapData(ctx context.Context, parent *api.KopsConfig, data []byte) error {
	// log := ctrl.LoggerFrom(ctx)

	clusterName := parent.Labels[clusterv1.ClusterNameLabel]

	if clusterName == "" {
		return fmt.Errorf("cluster name label %q not yet set", clusterv1.ClusterNameLabel)
	}

	secretName := types.NamespacedName{
		Namespace: parent.GetNamespace(),
		Name:      parent.GetName(),
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName.Name,
			Namespace: secretName.Namespace,
			Labels: map[string]string{
				clusterv1.ClusterNameLabel: clusterName,
			},
		},
		Data: map[string][]byte{
			"value": data,
			// "format": []byte(scope.Config.Spec.Format),
		},
		Type: clusterv1.ClusterSecretType,
	}

	parentAPIVersion, parentKind := parent.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()
	secret.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: parentAPIVersion,
			Kind:       parentKind,
			Name:       parent.GetName(),
			UID:        parent.GetUID(),
			Controller: pointer.Bool(true),
		},
	}

	var existing corev1.Secret
	if err := r.client.Get(ctx, secretName, &existing); err != nil {
		if apierrors.IsNotFound(err) {
			if err := r.client.Create(ctx, secret); err != nil {
				return fmt.Errorf("failed to create bootstrap data secret for KopsConfig %s/%s: %w", parent.GetNamespace(), parent.GetName(), err)
			}
		} else {
			return fmt.Errorf("failed to get bootstrap data secret: %w", err)
		}
	} else {
		// TODO: Verify that the existing secret "matches"
		klog.Warningf("TODO: verify that the existing secret matches our expected value")
	}

	parent.Status.DataSecretName = pointer.String(secret.Name)
	parent.Status.Ready = true
	// conditions.MarkTrue(scope.Config, bootstrapv1.DataSecretAvailableCondition)
	return nil
}

func (r *KopsConfigReconciler) buildBootstrapData(ctx context.Context) ([]byte, error) {
	// tf := &TemplateFunctions{
	// 	KopsModelContext: *modelContext,
	// 	cloud:            cloud,
	// }
	// TODO: Make dynamic
	clusterName := "clusterapi.k8s.local"
	clusterStoreBasePath := "gs://kops-state-justinsb-root-20220725"

	wellKnownAddresses := model.WellKnownAddresses{}
	wellKnownAddresses[wellknownservices.KopsController] = append(wellKnownAddresses[wellknownservices.KopsController], "10.0.16.2")
	wellKnownAddresses[wellknownservices.KubeAPIServer] = append(wellKnownAddresses[wellknownservices.KubeAPIServer], "10.0.16.2")

	vfsContext := vfs.NewVFSContext()
	basePath, err := vfsContext.BuildVfsPath(clusterStoreBasePath)
	if err != nil {
		return nil, fmt.Errorf("parsing vfs base path: %w", err)
	}

	// cluster := &kops.Cluster{}
	// cluster.Spec.KubernetesVersion = "1.28.3"
	// cluster.Spec.KubeAPIServer = &kops.KubeAPIServerConfig{}

	vfsClientset := vfsclientset.NewVFSClientset(vfsContext, basePath)
	cluster, err := vfsClientset.GetCluster(ctx, clusterName)
	if err != nil {
		return nil, fmt.Errorf("getting cluster %q: %w", clusterName, err)
	}

	if cluster.Spec.KubeAPIServer == nil {
		cluster.Spec.KubeAPIServer = &kops.KubeAPIServerConfig{}
	}

	ig := &kops.InstanceGroup{}
	ig.Spec.Role = kops.InstanceGroupRoleNode

	getAssets := false
	assetBuilder := assets.NewAssetBuilder(vfsContext, cluster.Spec.Assets, cluster.Spec.KubernetesVersion, getAssets)

	encryptionConfigSecretHash := ""
	// if fi.ValueOf(c.Cluster.Spec.EncryptionConfig) {
	// 	secret, err := secretStore.FindSecret("encryptionconfig")
	// 	if err != nil {
	// 		return fmt.Errorf("could not load encryptionconfig secret: %v", err)
	// 	}
	// 	if secret == nil {
	// 		fmt.Println("")
	// 		fmt.Println("You have encryptionConfig enabled, but no encryptionconfig secret has been set.")
	// 		fmt.Println("See `kops create secret encryptionconfig -h` and https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/")
	// 		return fmt.Errorf("could not find encryptionconfig secret")
	// 	}
	// 	hashBytes := sha256.Sum256(secret.Data)
	// 	encryptionConfigSecretHash = base64.URLEncoding.EncodeToString(hashBytes[:])
	// }

	nodeUpAssets := make(map[architectures.Architecture]*mirrors.MirroredAsset)
	for _, arch := range architectures.GetSupported() {
		asset, err := wellknownassets.NodeUpAsset(assetBuilder, arch)
		if err != nil {
			return nil, err
		}
		nodeUpAssets[arch] = asset
	}

	assets := make(map[architectures.Architecture][]*mirrors.MirroredAsset)
	configBuilder, err := cloudup.NewNodeUpConfigBuilder(cluster, assetBuilder, assets, encryptionConfigSecretHash)
	if err != nil {
		return nil, err
	}

	// bootstrapScript := &model.BootstrapScript{
	// 	// KopsModelContext:    modelContext,
	// 	Lifecycle: fi.LifecycleSync,
	// 	// NodeUpConfigBuilder: configBuilder,
	// 	// NodeUpAssets:        c.NodeUpAssets,
	// }

	keysets := make(map[string]*fi.Keyset)

	keystore, err := vfsClientset.KeyStore(cluster)
	if err != nil {
		return nil, err
	}

	for _, keyName := range []string{"kubernetes-ca"} {
		keyset, err := keystore.FindKeyset(ctx, keyName)
		if err != nil {
			return nil, fmt.Errorf("getting keyset %q: %w", keyName, err)
		}

		if keyset == nil {
			return nil, fmt.Errorf("failed to find keyset %q", keyName)
		}

		keysets[keyName] = keyset
	}

	_, bootConfig, err := configBuilder.BuildConfig(ig, wellKnownAddresses, keysets)
	if err != nil {
		return nil, err
	}

	// configData, err := utils.YamlMarshal(config)
	// if err != nil {
	// 	return nil, fmt.Errorf("error converting nodeup config to yaml: %v", err)
	// }
	// sum256 := sha256.Sum256(configData)
	// bootConfig.NodeupConfigHash = base64.StdEncoding.EncodeToString(sum256[:])
	// b.nodeupConfig.Resource = fi.NewBytesResource(configData)

	var nodeupScript resources.NodeUpScript
	nodeupScript.NodeUpAssets = nodeUpAssets
	nodeupScript.BootConfig = bootConfig

	{
		nodeupScript.EnvironmentVariables = func() (string, error) {
			env := make(map[string]string)

			// env, err := b.buildEnvironmentVariables()
			// if err != nil {
			// 	return "", err
			// }

			// Sort keys to have a stable sequence of "export xx=xxx"" statements
			var keys []string
			for k := range env {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			var b bytes.Buffer
			for _, k := range keys {
				b.WriteString(fmt.Sprintf("export %s=%s\n", k, env[k]))
			}
			return b.String(), nil
		}

		nodeupScript.ProxyEnv = func() (string, error) {
			return "", nil
			// return b.createProxyEnv(cluster.Spec.Networking.EgressProxy)
		}
	}

	// TODO: nodeupScript.CompressUserData = fi.ValueOf(b.ig.Spec.CompressUserData)

	// By setting some sysctls early, we avoid broken configurations that prevent nodeup download.
	// See https://github.com/kubernetes/kops/issues/10206 for details.
	// TODO: nodeupScript.SetSysctls = setSysctls()

	nodeupScript.CloudProvider = string(cluster.Spec.GetCloudProvider())

	nodeupScriptResource, err := nodeupScript.Build()
	if err != nil {
		return nil, err
	}

	b, err := fi.ResourceAsBytes(nodeupScriptResource)
	if err != nil {
		return nil, err
	}

	return b, nil
}
