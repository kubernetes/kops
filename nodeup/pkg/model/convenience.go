/*
Copyright 2016 The Kubernetes Authors.

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

package model

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// s is a helper that builds a *string from a string value
func s(v string) *string {
	return fi.String(v)
}

// i64 is a helper that builds a *int64 from an int64 value
func i64(v int64) *int64 {
	return fi.Int64(v)
}

// b returns a pointer to a boolean
func b(v bool) *bool {
	return fi.Bool(v)
}

// containsRole checks if a collection roles contains role v
func containsRole(v kops.InstanceGroupRole, list []kops.InstanceGroupRole) bool {
	for _, x := range list {
		if v == x {
			return true
		}
	}

	return false
}

// buildDockerEnvironmentVars just converts a series of keypairs to docker environment variables switches
func buildDockerEnvironmentVars(env map[string]string) []string {
	var list []string
	for k, v := range env {
		list = append(list, []string{"-e", fmt.Sprintf("%s=%s", k, v)}...)
	}

	return list
}

func getProxyEnvVars(proxies *kops.EgressProxySpec) []v1.EnvVar {
	if proxies == nil {
		glog.V(8).Info("proxies is == nil, returning empty list")
		return []v1.EnvVar{}
	}

	if proxies.HTTPProxy.Host == "" {
		glog.Warning("EgressProxy set but no proxy host provided")
	}

	var httpProxyURL string
	if proxies.HTTPProxy.Port == 0 {
		httpProxyURL = "http://" + proxies.HTTPProxy.Host
	} else {
		httpProxyURL = "http://" + proxies.HTTPProxy.Host + ":" + strconv.Itoa(proxies.HTTPProxy.Port)
	}

	noProxy := proxies.ProxyExcludes

	return []v1.EnvVar{
		{Name: "http_proxy", Value: httpProxyURL},
		{Name: "https_proxy", Value: httpProxyURL},
		{Name: "NO_PROXY", Value: noProxy},
		{Name: "no_proxy", Value: noProxy},
	}
}

// buildCertificateRequest retrieves the certificate from a keystore
func buildCertificateRequest(c *fi.ModelBuilderContext, b *NodeupModelContext, name, path string) error {
	cert, err := b.KeyStore.FindCert(name)
	if err != nil {
		return err
	}

	if cert == nil {
		return fmt.Errorf("certificate %q not found", name)
	}

	serialized, err := cert.AsString()
	if err != nil {
		return err
	}

	location := filepath.Join(b.PathSrvKubernetes(), fmt.Sprintf("%s.pem", name))
	if path != "" {
		location = path
	}

	c.AddTask(&nodetasks.File{
		Path:     location,
		Contents: fi.NewStringResource(serialized),
		Type:     nodetasks.FileType_File,
		Mode:     s("0600"),
	})

	return nil
}

// buildPrivateKeyRequest retrieves a private key from the store
func buildPrivateKeyRequest(c *fi.ModelBuilderContext, b *NodeupModelContext, name, path string) error {
	k, err := b.KeyStore.FindPrivateKey(name)
	if err != nil {
		return err
	}

	if k == nil {
		return fmt.Errorf("private key %q not found", name)
	}

	serialized, err := k.AsString()
	if err != nil {
		return err
	}

	location := filepath.Join(b.PathSrvKubernetes(), fmt.Sprintf("%s-key.pem", name))
	if path != "" {
		location = path
	}

	c.AddTask(&nodetasks.File{
		Path:     location,
		Contents: fi.NewStringResource(serialized),
		Type:     nodetasks.FileType_File,
		Mode:     s("0600"),
	})

	return nil
}

// sortedStrings is just a one liner helper methods
func sortedStrings(list []string) []string {
	sort.Strings(list)

	return list
}

// addHostPathMapping is shorthand for mapping a host path into a container
func addHostPathMapping(pod *v1.Pod, container *v1.Container, name, path string) *v1.VolumeMount {
	pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: path,
			},
		},
	})

	container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
		Name:      name,
		MountPath: path,
		ReadOnly:  true,
	})

	return &container.VolumeMounts[len(container.VolumeMounts)-1]
}

// convEtcdSettingsToMs converts etcd settings to a string rep of int milliseconds
func convEtcdSettingsToMs(dur *metav1.Duration) string {
	return strconv.FormatInt(dur.Nanoseconds()/1000000, 10)
}
