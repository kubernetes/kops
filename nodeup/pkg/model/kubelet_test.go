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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/distributions"
	"k8s.io/kops/util/pkg/vfs"
)

func Test_InstanceGroupKubeletMerge(t *testing.T) {
	cluster := &kops.Cluster{}
	cluster.Spec.Kubelet = &kops.KubeletConfigSpec{}
	cluster.Spec.Kubelet.NvidiaGPUs = 0
	cluster.Spec.KubernetesVersion = "1.6.0"

	instanceGroup := &kops.InstanceGroup{}
	instanceGroup.Spec.Kubelet = &kops.KubeletConfigSpec{}
	instanceGroup.Spec.Kubelet.NvidiaGPUs = 1
	instanceGroup.Spec.Role = kops.InstanceGroupRoleNode

	config, bootConfig := nodeup.NewConfig(cluster, instanceGroup)
	b := &KubeletBuilder{
		&NodeupModelContext{
			Cluster:      cluster,
			BootConfig:   bootConfig,
			NodeupConfig: config,
		},
	}
	if err := b.Init(); err != nil {
		t.Error(err)
	}

	mergedKubeletSpec, err := b.buildKubeletConfigSpec()
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
		input := testutils.BuildMinimalMasterInstanceGroup("eu-central-1a")
		input.Spec.Taints = g.taints

		ig, err := cloudup.PopulateInstanceGroupSpec(cluster, &input, nil, nil)
		if err != nil {
			t.Fatalf("failed to populate ig: %v", err)
		}

		config, bootConfig := nodeup.NewConfig(cluster, ig)
		b := &KubeletBuilder{
			&NodeupModelContext{
				Cluster:      cluster,
				BootConfig:   bootConfig,
				NodeupConfig: config,
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
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.18.0")
	h.SetupMockAWS()

	basedir := "tests/kubelet/featuregates"

	context := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	model, err := testutils.LoadModel(basedir)
	if err != nil {
		t.Fatal(err)
	}

	nodeUpModelContext, err := BuildNodeupModelContext(model)
	if err != nil {
		t.Fatalf("error loading model %q: %v", basedir, err)
		return
	}
	runKubeletBuilder(t, context, nodeUpModelContext)

	testutils.ValidateTasks(t, filepath.Join(basedir, "tasks.yaml"), context)
}

func Test_RunKubeletBuilderWarmPool(t *testing.T) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.18.0")
	h.SetupMockAWS()

	basedir := "tests/kubelet/warmpool"

	context := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	model, err := testutils.LoadModel(basedir)
	if err != nil {
		t.Fatal(err)
	}

	nodeUpModelContext, err := BuildNodeupModelContext(model)
	if err != nil {
		t.Fatalf("error loading model %q: %v", basedir, err)
		return
	}

	nodeUpModelContext.ConfigurationMode = "Warming"

	runKubeletBuilder(t, context, nodeUpModelContext)

	testutils.ValidateTasks(t, filepath.Join(basedir, "tasks.yaml"), context)
}

func runKubeletBuilder(t *testing.T, context *fi.ModelBuilderContext, nodeupModelContext *NodeupModelContext) {
	if err := nodeupModelContext.Init(); err != nil {
		t.Fatalf("error from nodeupModelContext.Init(): %v", err)
	}

	builder := KubeletBuilder{NodeupModelContext: nodeupModelContext}

	kubeletConfig, err := builder.buildKubeletConfigSpec()
	if err != nil {
		t.Fatalf("error from KubeletBuilder buildKubeletConfig: %v", err)
		return
	}
	{
		fileTask, err := buildKubeletComponentConfig(kubeletConfig)
		if err != nil {
			t.Fatalf("error from KubeletBuilder buildKubeletComponentConfig: %v", err)
			return
		}

		context.AddTask(fileTask)
	}
	{
		fileTask, err := builder.buildSystemdEnvironmentFile(kubeletConfig)
		if err != nil {
			t.Fatalf("error from KubeletBuilder buildSystemdEnvironmentFile: %v", err)
			return
		}
		context.AddTask(fileTask)
	}
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
}

func BuildNodeupModelContext(model *testutils.Model) (*NodeupModelContext, error) {
	if model.Cluster == nil {
		return nil, fmt.Errorf("no cluster found in model")
	}

	nodeupModelContext := &NodeupModelContext{
		Architecture:  "amd64",
		BootConfig:    &nodeup.BootConfig{},
		CloudProvider: model.Cluster.Spec.GetCloudProvider(),
		NodeupConfig: &nodeup.Config{
			CAs:        map[string]string{},
			KeypairIDs: map[string]string{},
		},
	}

	// Populate the cluster
	cloud, err := cloudup.BuildCloud(model.Cluster)
	if err != nil {
		return nil, fmt.Errorf("error from BuildCloud: %v", err)
	}

	err = cloudup.PerformAssignments(model.Cluster, cloud)
	if err != nil {
		return nil, fmt.Errorf("error from PerformAssignments: %v", err)
	}

	nodeupModelContext.Cluster, err = mockedPopulateClusterSpec(model.Cluster, cloud)
	if err != nil {
		return nil, fmt.Errorf("unexpected error from mockedPopulateClusterSpec: %v", err)
	}

	if len(model.InstanceGroups) == 0 {
		// We tolerate this - not all tests need an instance group
	} else if len(model.InstanceGroups) == 1 {
		nodeupModelContext.NodeupConfig, nodeupModelContext.BootConfig = nodeup.NewConfig(nodeupModelContext.Cluster, model.InstanceGroups[0])
	} else {
		return nil, fmt.Errorf("unexpected number of instance groups: found %d", len(model.InstanceGroups))
	}

	// Are we mocking out too much of the apply_cluster logic?
	nodeupModelContext.NodeupConfig.CAs["kubernetes-ca"] = dummyCertificate + nextCertificate
	nodeupModelContext.NodeupConfig.KeypairIDs["kubernetes-ca"] = "3"
	nodeupModelContext.NodeupConfig.KeypairIDs["service-account"] = "2"

	if nodeupModelContext.NodeupConfig.APIServerConfig != nil {
		saPublicKeys, _ := rotatingPrivateKeyset().ToPublicKeys()
		nodeupModelContext.NodeupConfig.APIServerConfig.ServiceAccountPublicKeys = saPublicKeys
	}

	nodeupModelContext.NodeupConfig.ContainerdConfig = nodeupModelContext.Cluster.Spec.Containerd
	updatePolicy := nodeupModelContext.Cluster.Spec.UpdatePolicy
	if updatePolicy == nil {
		updatePolicy = fi.String(kops.UpdatePolicyAutomatic)
	}
	nodeupModelContext.NodeupConfig.UpdatePolicy = *updatePolicy

	nodeupModelContext.NodeupConfig.KubeletConfig.PodManifestPath = "/etc/kubernetes/manifests"

	return nodeupModelContext, nil
}

func mockedPopulateClusterSpec(c *kops.Cluster, cloud fi.Cloud) (*kops.Cluster, error) {
	vfs.Context.ResetMemfsContext(true)

	assetBuilder := assets.NewAssetBuilder(c, false)
	basePath, err := vfs.Context.BuildVfsPath("memfs://tests")
	if err != nil {
		return nil, fmt.Errorf("error building vfspath: %v", err)
	}
	clientset := vfsclientset.NewVFSClientset(basePath)
	return cloudup.PopulateClusterSpec(clientset, c, cloud, assetBuilder)
}

// Fixed cert and key, borrowed from the create_kubecfg_test.go test
// Wouldn't actually work in a real environment, but good enough for (today's) tests

const (
	dummyCertificate    = "-----BEGIN CERTIFICATE-----\nMIIC2DCCAcCgAwIBAgIRALJXAkVj964tq67wMSI8oJQwDQYJKoZIhvcNAQELBQAw\nFTETMBEGA1UEAxMKa3ViZXJuZXRlczAeFw0xNzEyMjcyMzUyNDBaFw0yNzEyMjcy\nMzUyNDBaMBUxEzARBgNVBAMTCmt1YmVybmV0ZXMwggEiMA0GCSqGSIb3DQEBAQUA\nA4IBDwAwggEKAoIBAQDgnCkSmtnmfxEgS3qNPaUCH5QOBGDH/inHbWCODLBCK9gd\nXEcBl7FVv8T2kFr1DYb0HVDtMI7tixRVFDLgkwNlW34xwWdZXB7GeoFgU1xWOQSY\nOACC8JgYTQ/139HBEvgq4sej67p+/s/SNcw34Kk7HIuFhlk1rRk5kMexKIlJBKP1\nYYUYetsJ/QpUOkqJ5HW4GoetE76YtHnORfYvnybviSMrh2wGGaN6r/s4ChOaIbZC\nAn8/YiPKGIDaZGpj6GXnmXARRX/TIdgSQkLwt0aTDBnPZ4XvtpI8aaL8DYJIqAzA\nNPH2b4/uNylat5jDo0b0G54agMi97+2AUrC9UUXpAgMBAAGjIzAhMA4GA1UdDwEB\n/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBVGR2r\nhzXzRMU5wriPQAJScszNORvoBpXfZoZ09FIupudFxBVU3d4hV9StKnQgPSGA5XQO\nHE97+BxJDuA/rB5oBUsMBjc7y1cde/T6hmi3rLoEYBSnSudCOXJE4G9/0f8byAJe\nrN8+No1r2VgZvZh6p74TEkXv/l3HBPWM7IdUV0HO9JDhSgOVF1fyQKJxRuLJR8jt\nO6mPH2UX0vMwVa4jvwtkddqk2OAdYQvH9rbDjjbzaiW0KnmdueRo92KHAN7BsDZy\nVpXHpqo1Kzg7D3fpaXCf5si7lqqrdJVXH4JC72zxsPehqgi8eIuqOBkiDWmRxAxh\n8yGeRx9AbknHh4Ia\n-----END CERTIFICATE-----\n"
	dummyKey            = "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA4JwpEprZ5n8RIEt6jT2lAh+UDgRgx/4px21gjgywQivYHVxH\nAZexVb/E9pBa9Q2G9B1Q7TCO7YsUVRQy4JMDZVt+McFnWVwexnqBYFNcVjkEmDgA\ngvCYGE0P9d/RwRL4KuLHo+u6fv7P0jXMN+CpOxyLhYZZNa0ZOZDHsSiJSQSj9WGF\nGHrbCf0KVDpKieR1uBqHrRO+mLR5zkX2L58m74kjK4dsBhmjeq/7OAoTmiG2QgJ/\nP2IjyhiA2mRqY+hl55lwEUV/0yHYEkJC8LdGkwwZz2eF77aSPGmi/A2CSKgMwDTx\n9m+P7jcpWreYw6NG9BueGoDIve/tgFKwvVFF6QIDAQABAoIBAA0ktjaTfyrAxsTI\nBezb7Zr5NBW55dvuII299cd6MJo+rI/TRYhvUv48kY8IFXp/hyUjzgeDLunxmIf9\n/Zgsoic9Ol44/g45mMduhcGYPzAAeCdcJ5OB9rR9VfDCXyjYLlN8H8iU0734tTqM\n0V13tQ9zdSqkGPZOIcq/kR/pylbOZaQMe97BTlsAnOMSMKDgnftY4122Lq3GYy+t\nvpr+bKVaQZwvkLoSU3rECCaKaghgwCyX7jft9aEkhdJv+KlwbsGY6WErvxOaLWHd\ncuMQjGapY1Fa/4UD00mvrA260NyKfzrp6+P46RrVMwEYRJMIQ8YBAk6N6Hh7dc0G\n8Z6i1m0CgYEA9HeCJR0TSwbIQ1bDXUrzpftHuidG5BnSBtax/ND9qIPhR/FBW5nj\n22nwLc48KkyirlfIULd0ae4qVXJn7wfYcuX/cJMLDmSVtlM5Dzmi/91xRiFgIzx1\nAsbBzaFjISP2HpSgL+e9FtSXaaqeZVrflitVhYKUpI/AKV31qGHf04sCgYEA6zTV\n99Sb49Wdlns5IgsfnXl6ToRttB18lfEKcVfjAM4frnkk06JpFAZeR+9GGKUXZHqs\nz2qcplw4d/moCC6p3rYPBMLXsrGNEUFZqBlgz72QA6BBq3X0Cg1Bc2ZbK5VIzwkg\nST2SSux6ccROfgULmN5ZiLOtdUKNEZpFF3i3qtsCgYADT/s7dYFlatobz3kmMnXK\nsfTu2MllHdRys0YGHu7Q8biDuQkhrJwhxPW0KS83g4JQym+0aEfzh36bWcl+u6R7\nKhKj+9oSf9pndgk345gJz35RbPJYh+EuAHNvzdgCAvK6x1jETWeKf6btj5pF1U1i\nQ4QNIw/QiwIXjWZeubTGsQKBgQCbduLu2rLnlyyAaJZM8DlHZyH2gAXbBZpxqU8T\nt9mtkJDUS/KRiEoYGFV9CqS0aXrayVMsDfXY6B/S/UuZjO5u7LtklDzqOf1aKG3Q\ndGXPKibknqqJYH+bnUNjuYYNerETV57lijMGHuSYCf8vwLn3oxBfERRX61M/DU8Z\nworz/QKBgQDCTJI2+jdXg26XuYUmM4XXfnocfzAXhXBULt1nENcogNf1fcptAVtu\nBAiz4/HipQKqoWVUYmxfgbbLRKKLK0s0lOWKbYdVjhEm/m2ZU8wtXTagNwkIGoyq\nY/C1Lox4f1ROJnCjc/hfcOjcxX5M8A8peecHWlVtUPKTJgxQ7oMKcw==\n-----END RSA PRIVATE KEY-----\n"
	previousCertificate = "-----BEGIN CERTIFICATE-----\nMIIBZzCCARGgAwIBAgIBAjANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQDEw9zZXJ2\naWNlLWFjY291bnQwHhcNMjEwNTAyMjAzMDA2WhcNMzEwNTAyMjAzMDA2WjAaMRgw\nFgYDVQQDEw9zZXJ2aWNlLWFjY291bnQwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA\n2JbeF8dNwqfEKKD65aGlVs58fWkA0qZdVLKw8qATzRBJTi1nqbj2kAR4gyy/C8Mx\nouxva/om9d7Sq8Ka55T7+wIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0T\nAQH/BAUwAwEB/zAdBgNVHQ4EFgQUI5beFHueAGyT1pQ6UTOdbMfj3gQwDQYJKoZI\nhvcNAQELBQADQQBwPLO+Np8o6k3aNBGKE4JTCOs06X72OXNivkWWWP/9XGz6x4DI\nHPU65kbUn/pWXBUVVlpsKsdmWA2Bu8pd/vD+\n-----END CERTIFICATE-----\n"
	previousKey         = "-----BEGIN RSA PRIVATE KEY-----\nMIIBPQIBAAJBANiW3hfHTcKnxCig+uWhpVbOfH1pANKmXVSysPKgE80QSU4tZ6m4\n9pAEeIMsvwvDMaLsb2v6JvXe0qvCmueU+/sCAwEAAQJBAKt/gmpHqP3qA3u8RA5R\n2W6L360Z2Mnza1FmkI/9StCCkJGjuE5yDhxU4JcVnFyX/nMxm2ockEEQDqRSu7Oo\nxTECIQD2QsUsgFL4FnXWzTclySJ6ajE4Cte3gSDOIvyMNMireQIhAOEnsV8UaSI+\nZyL7NMLzMPLCgtsrPnlamr8gdrEHf9ITAiEAxCCLbpTI/4LL2QZZrINTLVGT34Fr\nKl/yI5pjrrp/M2kCIQDfOktQyRuzJ8t5kzWsUxCkntS+FxHJn1rtQ3Jp8dV4oQIh\nAOyiVWDyLZJvg7Y24Ycmp86BZjM9Wk/BfWpBXKnl9iDY\n-----END RSA PRIVATE KEY-----"
	nextCertificate     = "-----BEGIN CERTIFICATE-----\nMIIBZzCCARGgAwIBAgIBBDANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQDEw9zZXJ2\naWNlLWFjY291bnQwHhcNMjEwNTAyMjAzMjE3WhcNMzEwNTAyMjAzMjE3WjAaMRgw\nFgYDVQQDEw9zZXJ2aWNlLWFjY291bnQwXDANBgkqhkiG9w0BAQEFAANLADBIAkEA\no4Tridlsf4Yz3UAiup/scSTiG/OqxkUW3Fz7zGKvVcLeYj9GEIKuzoB1VFk1nboD\nq4cCuGLfdzaQdCQKPIsDuwIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0T\nAQH/BAUwAwEB/zAdBgNVHQ4EFgQUhPbxEmUbwVOCa+fZgxreFhf67UEwDQYJKoZI\nhvcNAQELBQADQQALMsyK2Q7C/bk27eCvXyZKUfrLvor10hEjwGhv14zsKWDeTj/J\nA1LPYp7U9VtFfgFOkVbkLE9Rstc0ltNrPqxA\n-----END CERTIFICATE-----\n"
	nextKey             = "-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBAKOE64nZbH+GM91AIrqf7HEk4hvzqsZFFtxc+8xir1XC3mI/RhCC\nrs6AdVRZNZ26A6uHArhi33c2kHQkCjyLA7sCAwEAAQJAejInjmEzqmzQr0NxcIN4\nPukwK3FBKl+RAOZfqNIKcww14mfOn7Gc6lF2zEC4GnLiB3tthbSXoBGi54nkW4ki\nyQIhANZNne9UhQlwyjsd3WxDWWrl6OOZ3J8ppMOIQni9WRLlAiEAw1XEdxPOSOSO\nB6rucpTT1QivVvyEFIb/ukvPm769Mh8CIQDNQwKnHdlfNX0+KljPPaMD1LrAZbr/\naC+8aWLhqtsKUQIgF7gUcTkwdV17eabh6Xv09Qtm7zMefred2etWvFy+8JUCIECv\nFYOKQVWHX+Q7CHX2K1oTECVnZuW1UItdDYVlFYxQ\n-----END RSA PRIVATE KEY-----"
)

func simplePrivateKeyset(cert, key string) *kops.Keyset {
	return &kops.Keyset{
		Spec: kops.KeysetSpec{
			PrimaryID: "3",
			Keys: []kops.KeysetItem{
				{
					Id:              "3",
					PublicMaterial:  []byte(cert),
					PrivateMaterial: []byte(key),
				},
			},
		},
	}
}

func rotatingPrivateKeyset() *fi.Keyset {
	keyset, _ := fi.NewKeyset(mustParseCertificate(previousCertificate), mustParseKey(previousKey))
	_, _ = keyset.AddItem(mustParseCertificate(nextCertificate), mustParseKey(nextKey), false)

	return keyset
}

func mustParseCertificate(s string) *pki.Certificate {
	k, err := pki.ParsePEMCertificate([]byte(s))
	if err != nil {
		klog.Fatalf("error parsing certificate %v", err)
	}
	return k
}

func mustParseKey(s string) *pki.PrivateKey {
	k, err := pki.ParsePEMPrivateKey([]byte(s))
	if err != nil {
		klog.Fatalf("error parsing private key %v", err)
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

	model, err := testutils.LoadModel(basedir)
	if err != nil {
		t.Fatal(err)
	}

	keystore := &fakeKeystore{}
	keystore.T = t
	saKeyset, _ := rotatingPrivateKeyset().ToAPIObject("service-account")
	keystore.privateKeysets = map[string]*kops.Keyset{
		"kubernetes-ca":           simplePrivateKeyset(dummyCertificate, dummyKey),
		"apiserver-aggregator-ca": simplePrivateKeyset(dummyCertificate, dummyKey),
		"kube-controller-manager": simplePrivateKeyset(dummyCertificate, dummyKey),
		"kube-proxy":              simplePrivateKeyset(dummyCertificate, dummyKey),
		"kube-scheduler":          simplePrivateKeyset(dummyCertificate, dummyKey),
		"service-account":         saKeyset,
	}

	nodeupModelContext, err := BuildNodeupModelContext(model)
	if err != nil {
		t.Fatalf("error loading model %q: %v", basedir, err)
	}

	nodeupModelContext.KeyStore = keystore

	nodeupModelContext.Distribution = distributions.DistributionUbuntu2004

	if err := nodeupModelContext.Init(); err != nil {
		t.Fatalf("error from nodeupModelContext.Init(): %v", err)
	}

	if err := builder(nodeupModelContext, context); err != nil {
		t.Fatalf("error from Build: %v", err)
	}

	testutils.ValidateTasks(t, filepath.Join(basedir, "tasks-"+key+".yaml"), context)
}

func Test_BuildComponentConfigFile(t *testing.T) {
	componentConfig := kops.KubeletConfigSpec{
		ShutdownGracePeriod:             &metav1.Duration{Duration: 30 * time.Second},
		ShutdownGracePeriodCriticalPods: &metav1.Duration{Duration: 10 * time.Second},
	}

	_, err := buildKubeletComponentConfig(&componentConfig)
	if err != nil {
		t.Errorf("Failed to build component config file: %v", err)
	}
}
