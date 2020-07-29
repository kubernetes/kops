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

package model

import (
	"fmt"
	"path/filepath"
	"testing"

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/vfs"
)

func Test_InstanceGroupKubeletMerge(t *testing.T) {
	var cluster = &kops.Cluster{}
	cluster.Spec.Kubelet = &kops.KubeletConfigSpec{}
	cluster.Spec.Kubelet.NvidiaGPUs = 0
	cluster.Spec.KubernetesVersion = "1.6.0"

	var instanceGroup = &kops.InstanceGroup{}
	instanceGroup.Spec.Kubelet = &kops.KubeletConfigSpec{}
	instanceGroup.Spec.Kubelet.NvidiaGPUs = 1
	instanceGroup.Spec.Role = kops.InstanceGroupRoleNode

	b := &KubeletBuilder{
		&NodeupModelContext{
			Cluster:       cluster,
			InstanceGroup: instanceGroup,
			NodeupConfig:  nodeup.NewConfig(cluster, instanceGroup),
		},
	}
	if err := b.Init(); err != nil {
		t.Error(err)
	}

	var mergedKubeletSpec, err = b.buildKubeletConfigSpec()
	if err != nil {
		t.Error(err)
	}
	if mergedKubeletSpec == nil {
		t.Error("Returned nil kubelet spec")
		t.FailNow()
	}

	if mergedKubeletSpec.NvidiaGPUs != instanceGroup.Spec.Kubelet.NvidiaGPUs {
		t.Errorf("InstanceGroup kubelet value (%d) should be reflected in merged output", instanceGroup.Spec.Kubelet.NvidiaGPUs)
	}
}

func TestTaintsApplied(t *testing.T) {
	tests := []struct {
		version           string
		taints            []string
		expectError       bool
		expectSchedulable bool
		expectTaints      []string
	}{
		{
			version:           "1.9.0",
			taints:            []string{"foo", "bar", "baz"},
			expectTaints:      []string{"foo", "bar", "baz"},
			expectSchedulable: true,
		},
	}

	for _, g := range tests {
		cluster := &kops.Cluster{Spec: kops.ClusterSpec{KubernetesVersion: g.version}}
		ig := &kops.InstanceGroup{Spec: kops.InstanceGroupSpec{Role: kops.InstanceGroupRoleMaster, Taints: g.taints}}

		b := &KubeletBuilder{
			&NodeupModelContext{
				Cluster:       cluster,
				InstanceGroup: ig,
				NodeupConfig:  nodeup.NewConfig(cluster, ig),
			},
		}
		if err := b.Init(); err != nil {
			t.Error(err)
		}

		c, err := b.buildKubeletConfigSpec()

		if g.expectError {
			if err == nil {
				t.Fatalf("Expected error but did not get one for version %q", g.version)
			}

			continue
		} else {
			if err != nil {
				t.Fatalf("Unexpected error for version %q: %v", g.version, err)
			}
		}

		if fi.BoolValue(c.RegisterSchedulable) != g.expectSchedulable {
			t.Fatalf("Expected RegisterSchedulable == %v, got %v (for %v)", g.expectSchedulable, fi.BoolValue(c.RegisterSchedulable), g.version)
		}

		if !stringSlicesEqual(g.expectTaints, c.Taints) {
			t.Fatalf("Expected taints %v, got %v", g.expectTaints, c.Taints)
		}
	}
}

func stringSlicesEqual(exp, other []string) bool {
	if exp == nil && other != nil {
		return false
	}

	if exp != nil && other == nil {
		return false
	}

	if len(exp) != len(other) {
		return false
	}

	for i, e := range exp {
		if other[i] != e {
			return false
		}
	}

	return true
}

func Test_RunKubeletBuilder(t *testing.T) {
	basedir := "tests/kubelet/featuregates"

	context := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}
	nodeUpModelContext, err := BuildNodeupModelContext(basedir)
	if err != nil {
		t.Fatalf("error loading model %q: %v", basedir, err)
		return
	}

	builder := KubeletBuilder{NodeupModelContext: nodeUpModelContext}

	kubeletConfig, err := builder.buildKubeletConfig()
	if err != nil {
		t.Fatalf("error from KubeletBuilder buildKubeletConfig: %v", err)
		return
	}

	fileTask, err := builder.buildSystemdEnvironmentFile(kubeletConfig)
	if err != nil {
		t.Fatalf("error from KubeletBuilder buildSystemdEnvironmentFile: %v", err)
		return
	}
	context.AddTask(fileTask)

	{
		task, err := builder.buildManifestDirectory(kubeletConfig)
		if err != nil {
			t.Fatalf("error from KubeletBuilder buildManifestDirectory: %v", err)
			return
		}
		context.AddTask(task)
	}

	{
		task := builder.buildSystemdService()
		if err != nil {
			t.Fatalf("error from KubeletBuilder buildSystemdService: %v", err)
			return
		}
		context.AddTask(task)
	}

	testutils.ValidateTasks(t, filepath.Join(basedir, "tasks.yaml"), context)
}

func BuildNodeupModelContext(basedir string) (*NodeupModelContext, error) {
	model, err := testutils.LoadModel(basedir)
	if err != nil {
		return nil, err
	}

	if model.Cluster == nil {
		return nil, fmt.Errorf("no cluster found in %s", basedir)
	}

	nodeUpModelContext := &NodeupModelContext{
		Cluster:      model.Cluster,
		Architecture: "amd64",
		NodeupConfig: &nodeup.Config{},
	}

	if len(model.InstanceGroups) == 0 {
		// We tolerate this - not all tests need an instance group
	} else if len(model.InstanceGroups) == 1 {
		nodeUpModelContext.InstanceGroup = model.InstanceGroups[0]
		nodeUpModelContext.NodeupConfig = nodeup.NewConfig(model.Cluster, nodeUpModelContext.InstanceGroup)
	} else {
		return nil, fmt.Errorf("unexpected number of instance groups in %s, found %d", basedir, len(model.InstanceGroups))
	}

	if err := nodeUpModelContext.Init(); err != nil {
		return nil, err
	}

	return nodeUpModelContext, nil
}

func mockedPopulateClusterSpec(c *kops.Cluster) (*kops.Cluster, error) {
	vfs.Context.ResetMemfsContext(true)

	assetBuilder := assets.NewAssetBuilder(c, "")
	basePath, err := vfs.Context.BuildVfsPath("memfs://tests")
	if err != nil {
		return nil, fmt.Errorf("error building vfspath: %v", err)
	}
	clientset := vfsclientset.NewVFSClientset(basePath)
	return cloudup.PopulateClusterSpec(clientset, c, assetBuilder)
}

// Fixed cert and key, borrowed from the create_kubecfg_test.go test
// Wouldn't actually work in a real environment, but good enough for (today's) tests

const dummyCertificate = "-----BEGIN CERTIFICATE-----\nMIIC2DCCAcCgAwIBAgIRALJXAkVj964tq67wMSI8oJQwDQYJKoZIhvcNAQELBQAw\nFTETMBEGA1UEAxMKa3ViZXJuZXRlczAeFw0xNzEyMjcyMzUyNDBaFw0yNzEyMjcy\nMzUyNDBaMBUxEzARBgNVBAMTCmt1YmVybmV0ZXMwggEiMA0GCSqGSIb3DQEBAQUA\nA4IBDwAwggEKAoIBAQDgnCkSmtnmfxEgS3qNPaUCH5QOBGDH/inHbWCODLBCK9gd\nXEcBl7FVv8T2kFr1DYb0HVDtMI7tixRVFDLgkwNlW34xwWdZXB7GeoFgU1xWOQSY\nOACC8JgYTQ/139HBEvgq4sej67p+/s/SNcw34Kk7HIuFhlk1rRk5kMexKIlJBKP1\nYYUYetsJ/QpUOkqJ5HW4GoetE76YtHnORfYvnybviSMrh2wGGaN6r/s4ChOaIbZC\nAn8/YiPKGIDaZGpj6GXnmXARRX/TIdgSQkLwt0aTDBnPZ4XvtpI8aaL8DYJIqAzA\nNPH2b4/uNylat5jDo0b0G54agMi97+2AUrC9UUXpAgMBAAGjIzAhMA4GA1UdDwEB\n/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBVGR2r\nhzXzRMU5wriPQAJScszNORvoBpXfZoZ09FIupudFxBVU3d4hV9StKnQgPSGA5XQO\nHE97+BxJDuA/rB5oBUsMBjc7y1cde/T6hmi3rLoEYBSnSudCOXJE4G9/0f8byAJe\nrN8+No1r2VgZvZh6p74TEkXv/l3HBPWM7IdUV0HO9JDhSgOVF1fyQKJxRuLJR8jt\nO6mPH2UX0vMwVa4jvwtkddqk2OAdYQvH9rbDjjbzaiW0KnmdueRo92KHAN7BsDZy\nVpXHpqo1Kzg7D3fpaXCf5si7lqqrdJVXH4JC72zxsPehqgi8eIuqOBkiDWmRxAxh\n8yGeRx9AbknHh4Ia\n-----END CERTIFICATE-----\n"
const dummyKey = "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA4JwpEprZ5n8RIEt6jT2lAh+UDgRgx/4px21gjgywQivYHVxH\nAZexVb/E9pBa9Q2G9B1Q7TCO7YsUVRQy4JMDZVt+McFnWVwexnqBYFNcVjkEmDgA\ngvCYGE0P9d/RwRL4KuLHo+u6fv7P0jXMN+CpOxyLhYZZNa0ZOZDHsSiJSQSj9WGF\nGHrbCf0KVDpKieR1uBqHrRO+mLR5zkX2L58m74kjK4dsBhmjeq/7OAoTmiG2QgJ/\nP2IjyhiA2mRqY+hl55lwEUV/0yHYEkJC8LdGkwwZz2eF77aSPGmi/A2CSKgMwDTx\n9m+P7jcpWreYw6NG9BueGoDIve/tgFKwvVFF6QIDAQABAoIBAA0ktjaTfyrAxsTI\nBezb7Zr5NBW55dvuII299cd6MJo+rI/TRYhvUv48kY8IFXp/hyUjzgeDLunxmIf9\n/Zgsoic9Ol44/g45mMduhcGYPzAAeCdcJ5OB9rR9VfDCXyjYLlN8H8iU0734tTqM\n0V13tQ9zdSqkGPZOIcq/kR/pylbOZaQMe97BTlsAnOMSMKDgnftY4122Lq3GYy+t\nvpr+bKVaQZwvkLoSU3rECCaKaghgwCyX7jft9aEkhdJv+KlwbsGY6WErvxOaLWHd\ncuMQjGapY1Fa/4UD00mvrA260NyKfzrp6+P46RrVMwEYRJMIQ8YBAk6N6Hh7dc0G\n8Z6i1m0CgYEA9HeCJR0TSwbIQ1bDXUrzpftHuidG5BnSBtax/ND9qIPhR/FBW5nj\n22nwLc48KkyirlfIULd0ae4qVXJn7wfYcuX/cJMLDmSVtlM5Dzmi/91xRiFgIzx1\nAsbBzaFjISP2HpSgL+e9FtSXaaqeZVrflitVhYKUpI/AKV31qGHf04sCgYEA6zTV\n99Sb49Wdlns5IgsfnXl6ToRttB18lfEKcVfjAM4frnkk06JpFAZeR+9GGKUXZHqs\nz2qcplw4d/moCC6p3rYPBMLXsrGNEUFZqBlgz72QA6BBq3X0Cg1Bc2ZbK5VIzwkg\nST2SSux6ccROfgULmN5ZiLOtdUKNEZpFF3i3qtsCgYADT/s7dYFlatobz3kmMnXK\nsfTu2MllHdRys0YGHu7Q8biDuQkhrJwhxPW0KS83g4JQym+0aEfzh36bWcl+u6R7\nKhKj+9oSf9pndgk345gJz35RbPJYh+EuAHNvzdgCAvK6x1jETWeKf6btj5pF1U1i\nQ4QNIw/QiwIXjWZeubTGsQKBgQCbduLu2rLnlyyAaJZM8DlHZyH2gAXbBZpxqU8T\nt9mtkJDUS/KRiEoYGFV9CqS0aXrayVMsDfXY6B/S/UuZjO5u7LtklDzqOf1aKG3Q\ndGXPKibknqqJYH+bnUNjuYYNerETV57lijMGHuSYCf8vwLn3oxBfERRX61M/DU8Z\nworz/QKBgQDCTJI2+jdXg26XuYUmM4XXfnocfzAXhXBULt1nENcogNf1fcptAVtu\nBAiz4/HipQKqoWVUYmxfgbbLRKKLK0s0lOWKbYdVjhEm/m2ZU8wtXTagNwkIGoyq\nY/C1Lox4f1ROJnCjc/hfcOjcxX5M8A8peecHWlVtUPKTJgxQ7oMKcw==\n-----END RSA PRIVATE KEY-----\n"

func mustParsePrivateKey(s string) *pki.PrivateKey {
	k, err := pki.ParsePEMPrivateKey([]byte(s))
	if err != nil {
		klog.Fatalf("error parsing private key %v", err)
	}
	return k
}

func mustParseCertificate(s string) *pki.Certificate {
	k, err := pki.ParsePEMCertificate([]byte(s))
	if err != nil {
		klog.Fatalf("error parsing certificate %v", err)
	}
	return k
}

func RunGoldenTest(t *testing.T, basedir string, key string, builder func(*NodeupModelContext, *fi.ModelBuilderContext) error) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.18.0")
	h.SetupMockAWS()

	context := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}
	nodeupModelContext, err := BuildNodeupModelContext(basedir)
	if err != nil {
		t.Fatalf("error loading model %q: %v", basedir, err)
	}

	keystore := &fakeCAStore{}
	keystore.T = t
	keystore.privateKeys = map[string]*pki.PrivateKey{
		"ca":                      mustParsePrivateKey(dummyKey),
		"apiserver-aggregator-ca": mustParsePrivateKey(dummyKey),
		"kube-controller-manager": mustParsePrivateKey(dummyKey),
		"kube-proxy":              mustParsePrivateKey(dummyKey),
		"kube-scheduler":          mustParsePrivateKey(dummyKey),
		"master":                  mustParsePrivateKey(dummyKey),
	}
	keystore.certs = map[string]*pki.Certificate{
		"ca":                      mustParseCertificate(dummyCertificate),
		"apiserver-aggregator-ca": mustParseCertificate(dummyCertificate),
		"kube-controller-manager": mustParseCertificate(dummyCertificate),
		"kube-proxy":              mustParseCertificate(dummyCertificate),
		"kube-scheduler":          mustParseCertificate(dummyCertificate),
	}

	nodeupModelContext.KeyStore = keystore

	// Populate the cluster
	{
		err := cloudup.PerformAssignments(nodeupModelContext.Cluster)
		if err != nil {
			t.Fatalf("error from PerformAssignments: %v", err)
		}

		full, err := mockedPopulateClusterSpec(nodeupModelContext.Cluster)
		if err != nil {
			t.Fatalf("unexpected error from mockedPopulateClusterSpec: %v", err)
		}
		nodeupModelContext.Cluster = full
	}

	if err := builder(nodeupModelContext, context); err != nil {
		t.Fatalf("error from Build: %v", err)
	}

	testutils.ValidateTasks(t, filepath.Join(basedir, "tasks-"+key+".yaml"), context)
}
