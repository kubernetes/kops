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

package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	certmanager "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const AnnotationPrefix = "addons.k8s.io/"

type Channel struct {
	Namespace string
	Name      string
}

// CurrentSystemGeneration holds our current SystemGeneration value.
// Version history:
//
//	0  Pre-history (and the default value); versions prior to prune.
//	1  Prune functionality introduced.
const CurrentSystemGeneration = 1

type ChannelVersion struct {
	Channel      *string `json:"channel,omitempty"`
	Id           string  `json:"id,omitempty"`
	ManifestHash string  `json:"manifestHash,omitempty"`

	// SystemGeneration holds the generation of the channels functionality.
	// It is used so that we reapply when we introduce new features, such as prune.
	SystemGeneration int `json:"systemGeneration,omitempty"`
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (c *ChannelVersion) String() string {
	s := "Channel=" + stringValue(c.Channel)
	if c.Id != "" {
		s += " Id=" + c.Id
	}
	if c.ManifestHash != "" {
		s += " ManifestHash=" + c.ManifestHash
	}
	s += " SystemGeneration=" + strconv.Itoa(c.SystemGeneration)
	return s
}

func ParseChannelVersion(s string) (*ChannelVersion, error) {
	v := &ChannelVersion{}
	err := json.Unmarshal([]byte(s), v)
	if err != nil {
		return nil, fmt.Errorf("error parsing version spec %q", s)
	}
	return v, nil
}

func FindChannelVersions(ns *v1.Namespace) map[string]*ChannelVersion {
	addons := make(map[string]*ChannelVersion)
	for k, v := range ns.Annotations {
		if !strings.HasPrefix(k, AnnotationPrefix) {
			continue
		}

		channelVersion, err := ParseChannelVersion(v)
		if err != nil {
			klog.Warningf("failed to parse annotation %q=%q", k, v)
			continue
		}

		name := strings.TrimPrefix(k, AnnotationPrefix)
		addons[name] = channelVersion
	}
	return addons
}

func (c *ChannelVersion) Encode() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("error encoding version spec: %v", err)
	}
	return string(data), nil
}

func (c *Channel) AnnotationName() string {
	return AnnotationPrefix + c.Name
}

func (c *ChannelVersion) replaces(name string, existing *ChannelVersion) bool {
	klog.V(6).Infof("Checking existing config for %q: %v compared to new channel: %v", name, existing, c)

	if c.Id != existing.Id {
		klog.V(4).Infof("cluster has different ids for %q (%q vs %q); will replace", name, c.Id, existing.Id)
		return true
	}

	if c.ManifestHash != existing.ManifestHash {
		klog.V(4).Infof("cluster has different ManifestHash for %q (%q vs %q); will replace", name, c.ManifestHash, existing.ManifestHash)
		return true
	}

	if existing.SystemGeneration != c.SystemGeneration {
		if existing.SystemGeneration > c.SystemGeneration {
			klog.V(4).Infof("cluster has newer SystemGeneration for %q (%v vs %v), will not replace", name, existing.SystemGeneration, c.SystemGeneration)
			return false
		} else {
			klog.V(4).Infof("cluster has different SystemGeneration for %q (%v vs %v); will replace", name, existing.SystemGeneration, c.SystemGeneration)
			return true
		}
	}

	klog.V(4).Infof("manifest Match for %q: %v", name, existing)
	return false
}

func (c *Channel) GetInstalledVersion(ctx context.Context, k8sClient kubernetes.Interface) (*ChannelVersion, error) {
	ns, err := k8sClient.CoreV1().Namespaces().Get(ctx, c.Namespace, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error querying namespace %q: %v", c.Namespace, err)
	}

	annotationValue, ok := ns.Annotations[c.AnnotationName()]
	if !ok {
		return nil, nil
	}

	return ParseChannelVersion(annotationValue)
}

func (c *Channel) IsPKIInstalled(ctx context.Context, k8sClient kubernetes.Interface, cmClient certmanager.Interface) (bool, error) {
	_, err := k8sClient.CoreV1().Secrets("kube-system").Get(ctx, c.Name+"-ca", metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return true, err
	}

	_, err = cmClient.CertmanagerV1().Issuers("kube-system").Get(ctx, c.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return true, err
	}

	return true, nil
}

type annotationPatch struct {
	Metadata annotationPatchMetadata `json:"metadata,omitempty"`
}

type annotationPatchMetadata struct {
	Annotations map[string]string `json:"annotations,omitempty"`
}

func (c *Channel) SetInstalledVersion(ctx context.Context, k8sClient kubernetes.Interface, version *ChannelVersion) error {
	// Primarily to check it exists
	_, err := k8sClient.CoreV1().Namespaces().Get(ctx, c.Namespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error querying namespace %q: %v", c.Namespace, err)
	}

	value, err := version.Encode()
	if err != nil {
		return err
	}

	annotationPatch := &annotationPatch{Metadata: annotationPatchMetadata{Annotations: map[string]string{c.AnnotationName(): value}}}
	annotationPatchJSON, err := json.Marshal(annotationPatch)
	if err != nil {
		return fmt.Errorf("error building annotation patch: %v", err)
	}

	klog.V(2).Infof("sending patch: %q", string(annotationPatchJSON))

	_, err = k8sClient.CoreV1().Namespaces().Patch(ctx, c.Namespace, types.StrategicMergePatchType, annotationPatchJSON, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("error applying annotation to namespace: %v", err)
	}
	return nil
}
