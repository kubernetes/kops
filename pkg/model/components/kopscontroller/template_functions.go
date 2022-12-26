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

package kopscontroller

import (
	"text/template"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model/components/etcdmanager"
	"k8s.io/kops/pkg/wellknownports"
)

// AddTemplateFunctions registers template functions for KopsController
func AddTemplateFunctions(cluster *kops.Cluster, dest template.FuncMap) {
	t := &templateFunctions{
		Cluster: cluster,
	}
	dest["KopsController"] = func() *templateFunctions {
		return t
	}
}

// templateFunctions implements the KopsController template object helper.
type templateFunctions struct {
	Cluster *kops.Cluster
}

// KopsControllerConfig returns the yaml configuration for kops-controller
func (t *templateFunctions) GossipServices() ([]*corev1.Service, error) {
	if !t.Cluster.IsGossip() {
		return nil, nil
	}

	var services []*corev1.Service

	// api service
	{
		service := buildHeadlessService(types.NamespacedName{Name: "api-internal", Namespace: "kube-system"})
		service.Spec.Ports = []corev1.ServicePort{
			{Name: "https", Port: 443, Protocol: corev1.ProtocolTCP},
		}
		service.Spec.Selector = map[string]string{
			"k8s-app": "kops-controller",
		}
		service.Labels = map[string]string{
			kops.DiscoveryLabelKey: "api",
		}
		services = append(services, service)
	}

	// kops-controller service
	{
		service := buildHeadlessService(types.NamespacedName{Name: "kops-controller-internal", Namespace: "kube-system"})
		service.Spec.Ports = []corev1.ServicePort{
			{Name: "https", Port: wellknownports.KopsControllerPort, Protocol: corev1.ProtocolTCP},
		}
		service.Spec.Selector = map[string]string{
			"k8s-app": "kops-controller",
		}
		service.Labels = map[string]string{
			kops.DiscoveryLabelKey: "kops-controller",
		}
		services = append(services, service)
	}

	// etcd services
	if featureflag.APIServerNodes.Enabled() {
		for _, etcdCluster := range t.Cluster.Spec.EtcdClusters {
			name := "etcd-" + etcdCluster.Name + "-internal"
			service := buildHeadlessService(types.NamespacedName{Name: name, Namespace: "kube-system"})
			ports, err := etcdmanager.PortsForCluster(etcdCluster)
			if err != nil {
				return nil, err
			}
			service.Spec.Ports = []corev1.ServicePort{
				{Name: "https", Port: int32(ports.ClientPort), Protocol: corev1.ProtocolTCP},
			}
			service.Labels = map[string]string{
				kops.DiscoveryLabelKey: etcdCluster.Name + ".etcd",
			}
			service.Spec.Selector = etcdmanager.SelectorForCluster(etcdCluster)
			services = append(services, service)
		}
	}

	// We set the target port, to make applying cleaner
	for _, service := range services {
		for i := range service.Spec.Ports {
			port := &service.Spec.Ports[i]
			if port.TargetPort == intstr.FromInt(0) {
				port.TargetPort = intstr.FromInt(int(port.Port))
			}
		}
	}

	return services, nil
}

// buildHeadlessService is a helper to build a headless service
func buildHeadlessService(name types.NamespacedName) *corev1.Service {
	s := &corev1.Service{}
	s.APIVersion = "v1"
	s.Kind = "Service"
	s.Name = name.Name
	s.Namespace = name.Namespace

	s.Spec.ClusterIP = corev1.ClusterIPNone
	s.Spec.Type = corev1.ServiceTypeClusterIP

	return s
}
