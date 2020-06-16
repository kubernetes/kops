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

package kubeapiserver

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
)

// KubeApiserverBuilder builds the static manifest for kube-apiserver-healthcheck sidecar
type KubeApiserverBuilder struct {
	*model.KopsModelContext
	Lifecycle    *fi.Lifecycle
	AssetBuilder *assets.AssetBuilder
}

var _ fi.ModelBuilder = &KubeApiserverBuilder{}

func (b *KubeApiserverBuilder) useHealthCheckSidecar(c *fi.ModelBuilderContext) bool {
	// Should we use our health-check proxy, which allows us to
	// query the secure port without enabling anonymous auth?
	useHealthCheckSidecar := true
	// We only turn on the proxy in k8s 1.17 and above
	if b.IsKubernetesLT("1.17") {
		useHealthCheckSidecar = false
	}

	return useHealthCheckSidecar
}

// Build creates the tasks relating to kube-apiserver
// Currently we only build the kube-apiserver-healthcheck sidecar
func (b *KubeApiserverBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.useHealthCheckSidecar(c) {
		return nil
	}

	manifest, err := b.buildManifest()
	if err != nil {
		return err
	}

	manifestYAML, err := k8scodecs.ToVersionedYaml(manifest)
	if err != nil {
		return fmt.Errorf("error marshaling manifest to yaml: %v", err)
	}

	key := "kube-apiserver-healthcheck"
	location := "manifests/static/" + key + ".yaml"

	c.AddTask(&fitasks.ManagedFile{
		Contents:  fi.WrapResource(fi.NewBytesResource(manifestYAML)),
		Lifecycle: b.Lifecycle,
		Location:  fi.String(location),
		Name:      fi.String("manifests-static-" + key),
	})

	b.AssetBuilder.StaticManifests = append(b.AssetBuilder.StaticManifests, &assets.StaticManifest{
		Key:   key,
		Path:  location,
		Roles: []kops.InstanceGroupRole{kops.InstanceGroupRoleMaster},
	})
	return nil
}

func (b *KubeApiserverBuilder) buildManifest() (*corev1.Pod, error) {
	return b.buildHealthcheckSidecar()
}

const defaultManifest = `
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: healthcheck
    image: kope/kube-apiserver-healthcheck:1.18.0-alpha.3
    livenessProbe:
      httpGet:
        # The sidecar serves a healthcheck on the same port,
        # but with a .kube-apiserver-healthcheck prefix
        path: /.kube-apiserver-healthcheck/healthz
        port: 8080
        host: 127.0.0.1
      initialDelaySeconds: 5
      timeoutSeconds: 5
    command:
    - /usr/bin/kube-apiserver-healthcheck
    args:
    - --ca-cert=/secrets/ca.crt
    - --client-cert=/secrets/client.crt
    - --client-key=/secrets/client.key
    volumeMounts:
    - name: healthcheck-secrets
      mountPath: /secrets
      readOnly: true
  volumes:
  - name: healthcheck-secrets
    hostPath:
      path: /etc/kubernetes/kube-apiserver-healthcheck/secrets
      type: Directory
`

// buildHealthcheckSidecar builds the partial pod for the healthcheck sidecar.
// nodeup will merge it into the kube-apiserver pod.
func (b *KubeApiserverBuilder) buildHealthcheckSidecar() (*corev1.Pod, error) {
	// TODO: pull from bundle
	bundle := "(embedded kube-apiserver-healthcheck manifest)"
	manifest := []byte(defaultManifest)

	var pod *corev1.Pod
	var container *corev1.Container
	{
		objects, err := model.ParseManifest(manifest)
		if err != nil {
			return nil, err
		}
		if len(objects) != 1 {
			return nil, fmt.Errorf("expected exactly one object in manifest %s, found %d", bundle, len(objects))
		}
		if podObject, ok := objects[0].(*corev1.Pod); !ok {
			return nil, fmt.Errorf("expected Pod object in manifest %s, found %T", bundle, objects[0])
		} else {
			pod = podObject
		}

		if len(pod.Spec.Containers) != 1 {
			return nil, fmt.Errorf("expected exactly one container in etcd-manager Pod, found %d", len(pod.Spec.Containers))
		}
		container = &pod.Spec.Containers[0]
	}

	// Remap image via AssetBuilder
	{
		remapped, err := b.AssetBuilder.RemapImage(container.Image)
		if err != nil {
			return nil, fmt.Errorf("unable to remap container image %q: %v", container.Image, err)
		}
		container.Image = remapped
	}

	return pod, nil
}
