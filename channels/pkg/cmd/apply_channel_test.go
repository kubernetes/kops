/*
Copyright 2022 The Kubernetes Authors.

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

package cmd

import (
	"context"
	"testing"

	cmfake "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	"k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/channels/pkg/channels"
	"k8s.io/kops/upup/pkg/fi"
)

func TestGetUpdates(t *testing.T) {
	// This test checks checks that the addon is applied to the correct namespace.
	// It should be applied to kube-system even though the same addon has already been applied to default.

	kubeSystemNS := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
		},
	}
	defaultNS := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
			Annotations: map[string]string{
				"addons.k8s.io/aws-ebs-csi-driver.addons.k8s.io": "{\"channel\":\"s3://mystatestore/cluster.example.com/addons/bootstrap-channel.yaml\",\"id\":\"k8s-1.17\",\"manifestHash\":\"abc\",\"systemGeneration\":1}",
			},
		},
	}
	k8sClient := fakek8s.NewSimpleClientset(&kubeSystemNS, &defaultNS)
	ctx := context.Background()

	channelVersions, err := getChannelVersions(ctx, k8sClient)
	if err != nil {
		t.Errorf("failed to get channel versions: %v", err)
	}

	menu := channels.NewAddonMenu()
	menu.Addons = map[string]*channels.Addon{
		"aws-ebs-csi-driver.addons.k8s.io": {
			Name: "aws-ebs-csi-driver.addons.k8s.io",
			Spec: &api.AddonSpec{
				Name:         fi.String("aws-ebs-csi-driver.addons.k8s.io"),
				Id:           "k8s-1.17",
				ManifestHash: "abc",
			},
		},
	}
	_, needUpdates, err := getUpdates(ctx, menu, k8sClient, cmfake.NewSimpleClientset(), channelVersions)
	if err != nil {
		t.Errorf("failed to get updates: %v", err)
	}

	if len(needUpdates) != 1 {
		t.Fatalf("expected 1 update, but got %d", len(needUpdates))
	}
	if needUpdates[0].GetNamespace() != "kube-system" {
		t.Errorf("expected update in kube-system, but update applied to %q", needUpdates[0].GetNamespace())
	}
}
