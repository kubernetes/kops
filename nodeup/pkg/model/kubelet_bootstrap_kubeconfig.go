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

package model

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	kubeconfigv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"k8s.io/klog"
	"k8s.io/kops/cmd/kops-controller/pkg/nodebootstrap/client"
	pb "k8s.io/kops/pkg/proto/nodebootstrap"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
)

// KubeletBootstrapKubeconfigBuilder is responsible for node authorization
type KubeletBootstrapKubeconfigBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubeletBootstrapKubeconfigBuilder{}

// Build creates the kubelet bootstrap kubeconfig, by talking to kops controller
// Master nodes (and configurations not using kops-controller here) need to self-bootstrap
func (b *KubeletBootstrapKubeconfigBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.UseKopsControllerForKubeletBootstrap() {
		// We're not using kops-controller for kubelet bootstrap
		return nil
	}

	if b.IsMaster {
		// We are a master, we need to self-bootstrap
		return nil
	}

	config := client.Options{}
	config.PopulateDefaults()

	config.Server = fmt.Sprintf("%s:%d", b.Cluster.Spec.MasterInternalName, wellknownports.KopsControllerGRPCPort)
	caCert, err := b.CACertificate()
	if err != nil {
		return err
	}
	config.CACertificates = caCert

	apiserverURL := fmt.Sprintf("https://%s", b.Cluster.Spec.MasterInternalName)

	c.AddTask(&KubeletBootstrapKubeconfigTask{
		Name:                    "kubelet-bootstrap",
		Path:                    b.KubeletBootstrapKubeconfig(),
		ClientConfig:            config,
		KubeconfigAPIServer:     apiserverURL,
		KubeconfigCACertificate: caCert,
	})

	return nil
}

// makeKubeconfig is responsible for generating a bootstrap kubeconfig
func makeKubeconfig(apiserverURL string, apiserverCACertificate []byte, token pb.Token) ([]byte, error) {
	name := "bootstrap-context"
	clusterName := "cluster"

	cfg := &kubeconfigv1.Config{
		APIVersion: "v1",
		Kind:       "Config",
		AuthInfos: []kubeconfigv1.NamedAuthInfo{
			{
				Name: name,
				AuthInfo: kubeconfigv1.AuthInfo{
					Token: token.BearerToken,
				},
			},
		},
		Clusters: []kubeconfigv1.NamedCluster{
			{
				Name: clusterName,
				Cluster: kubeconfigv1.Cluster{
					Server:                   apiserverURL,
					CertificateAuthorityData: apiserverCACertificate,
				},
			},
		},
		Contexts: []kubeconfigv1.NamedContext{
			{
				Name: name,
				Context: kubeconfigv1.Context{
					Cluster:  clusterName,
					AuthInfo: name,
				},
			},
		},
		CurrentContext: name,
	}

	return json.MarshalIndent(cfg, "", "  ")
}

type KubeletBootstrapKubeconfigTask struct {
	Name string

	Path string

	ClientConfig client.Options

	// KubeconfigAPIServer is the kubeconfig API server
	KubeconfigAPIServer string

	// KubeconfigCACertificate is the kubeconfig API server ca certificate
	KubeconfigCACertificate []byte
}

var _ fi.HasDependencies = &KubeletBootstrapKubeconfigTask{}
var _ fi.HasName = &KubeletBootstrapKubeconfigTask{}

func (p *KubeletBootstrapKubeconfigTask) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	for _, v := range tasks {
		switch v.(type) {
		default:
			klog.Warningf("Unhandled type %T in KubeletBootstrapKubeconfigTask::GetDependencies: %v", v, v)
			//deps = append(deps, v)
		}
	}
	return deps
}

func (s *KubeletBootstrapKubeconfigTask) String() string {
	return fmt.Sprintf("KubeletBootstrapKubeconfigTask: %s", s.Name)
}

func (e *KubeletBootstrapKubeconfigTask) Find(c *fi.Context) (*KubeletBootstrapKubeconfigTask, error) {
	_, err := ioutil.ReadFile(e.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("error reading file %q: %v", e.Path, err)
	}

	// We assume it matches - we don't have any reconfiguration today
	actual := *e
	return &actual, nil
}

func (e *KubeletBootstrapKubeconfigTask) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *KubeletBootstrapKubeconfigTask) CheckChanges(a, e, changes *KubeletBootstrapKubeconfigTask) error {
	return nil
}

func (_ *KubeletBootstrapKubeconfigTask) RenderLocal(t *local.LocalTarget, a, e, changes *KubeletBootstrapKubeconfigTask) error {
	ctx := context.TODO()

	p := e.Path

	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, os.FileMode(0770)); err != nil {
		return fmt.Errorf("error creating directories %q: %v", dir, err)
	}

	client, err := client.New(ctx, &e.ClientConfig)
	if err != nil {
		return fmt.Errorf("error to build node bootstrap client: %v", err)
	}

	token, err := client.CreateKubeletBootstrapToken(ctx)
	if err != nil {
		return fmt.Errorf("unable to get kubelet bootstrap token: %v", err)
	}

	kubeconfig, err := makeKubeconfig(e.KubeconfigAPIServer, e.KubeconfigCACertificate, token)
	if err != nil {
		return fmt.Errorf("error building bootstrap kubeconfig: %v", err)
	}

	return ioutil.WriteFile(p, kubeconfig, os.FileMode(0640))
}

func (_ *KubeletBootstrapKubeconfigTask) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *KubeletBootstrapKubeconfigTask) error {
	return fmt.Errorf("RenderCloudInit not supported for KubeletBootstrapKubeconfig")
}

var _ fi.HasName = &KubeletBootstrapKubeconfigTask{}

func (f *KubeletBootstrapKubeconfigTask) GetName() *string {
	return &f.Name
}

func (f *KubeletBootstrapKubeconfigTask) SetName(name string) {
	klog.Fatalf("SetName not supported for KubeletBootstrapKubeconfigTask task")
}
