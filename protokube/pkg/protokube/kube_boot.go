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

package protokube

import (
	"fmt"
	"net"
	"os/exec"
	"time"

	"github.com/golang/glog"
)

var (
	// Containerized indicates the etcd is containerized
	Containerized = false
	// RootFS is the root fs path
	RootFS = "/"
)

// KubeBoot is the options for the protokube service
type KubeBoot struct {
	// Channels is a list of channel to apply
	Channels []string
	// InitializeRBAC should be set to true if we should create the core RBAC roles
	InitializeRBAC bool
	// InternalDNSSuffix is the dns zone we are living in
	InternalDNSSuffix string
	// InternalIP is the internal ip address of the node
	InternalIP net.IP
	// ApplyTaints controls whether we set taints based on the master label
	ApplyTaints bool
	// DNS is the dns provider
	DNS DNSProvider
	// ModelDir is the model directory
	ModelDir string
	// Etcd container registry location.
	EtcdImageSource string
	// TLSCA is the path to a client ca for etcd
	TLSCA string
	// TLSCert is the path to a tls certificate for etcd
	TLSCert string
	// TLSKey is the path to a tls private key for etcd
	TLSKey string
	// PeerCA is the path to a peer ca for etcd
	PeerCA string
	// PeerCert is the path to a peer certificate for etcd
	PeerCert string
	// PeerKey is the path to a peer private key for etcd
	PeerKey string
	// Kubernetes is the context methods for kubernetes
	Kubernetes *KubernetesContext
	// Master indicates we are a master node
	Master          bool
	volumeMounter   *VolumeMountController
	etcdControllers map[string]*EtcdController
}

// Init is responsible for initializing the controllers
func (k *KubeBoot) Init(volumesProvider Volumes) {
	k.volumeMounter = newVolumeMountController(volumesProvider)
	k.etcdControllers = make(map[string]*EtcdController)
}

// RunSyncLoop is responsible for provision the cluster
func (k *KubeBoot) RunSyncLoop() {
	for {
		if err := k.syncOnce(); err != nil {
			glog.Warningf("error during attempt to bootstrap (will sleep and retry): %v", err)
		}

		time.Sleep(1 * time.Minute)
	}
}

func (k *KubeBoot) syncOnce() error {
	if k.Master {
		// attempt to mount the volumes
		volumes, err := k.volumeMounter.mountMasterVolumes()
		if err != nil {
			return err
		}

		for _, v := range volumes {
			for _, etcdSpec := range v.Info.EtcdClusters {
				key := etcdSpec.ClusterKey + "::" + etcdSpec.NodeName
				etcdController := k.etcdControllers[key]
				if etcdController == nil {
					glog.Infof("Found etcd cluster spec on volume %q: %v", v.ID, etcdSpec)
					etcdController, err := newEtcdController(k, v, etcdSpec)
					if err != nil {
						glog.Warningf("error building etcd controller: %v", err)
					} else {
						k.etcdControllers[key] = etcdController
						go etcdController.RunSyncLoop()
					}
				}
			}
		}
	} else {
		glog.V(4).Infof("Not in role master; won't scan for volumes")
	}

	if k.Master && k.ApplyTaints {
		if err := applyMasterTaints(k.Kubernetes); err != nil {
			glog.Warningf("error updating master taints: %v", err)
		}
	}

	if k.InitializeRBAC {
		// @TODO: Idempotency: good question; not sure this should ever be done on the node though
		if err := applyRBAC(k.Kubernetes); err != nil {
			glog.Warningf("error initializing rbac: %v", err)
		}
	}

	// Ensure kubelet is running. We avoid doing this automatically so
	// that when kubelet comes up the first time, all volume mounts
	// and DNS are available, avoiding the scenario where
	// etcd/apiserver retry too many times and go into backoff.
	if err := startKubeletService(); err != nil {
		glog.Warningf("error ensuring kubelet started: %v", err)
	}

	for _, channel := range k.Channels {
		if err := applyChannel(channel); err != nil {
			glog.Warningf("error applying channel %q: %v", channel, err)
		}
	}

	return nil
}

// startKubeletService is responsible for checking and if not starting the kubelet service
func startKubeletService() error {
	// TODO: Check/log status of kubelet
	// (in particular, we want to avoid kubernetes/kubernetes#40123 )
	glog.V(2).Infof("ensuring that kubelet systemd service is running")

	cmd := exec.Command("systemctl", "status", "--no-block", "kubelet")
	output, err := cmd.CombinedOutput()
	glog.V(2).Infof("'systemctl status kubelet' output:\n%s", string(output))
	if err == nil {
		glog.V(2).Infof("kubelet systemd service already running")
		return nil
	}

	glog.Infof("kubelet systemd service not running. Starting")
	cmd = exec.Command("systemctl", "start", "--no-block", "kubelet")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error starting kubelet: %v\nOutput: %s", err, output)
	}
	glog.V(2).Infof("'systemctl start kubelet' output:\n%s", string(output))

	return nil
}

func pathFor(hostPath string) string {
	if hostPath[0] != '/' {
		glog.Fatalf("path was not absolute: %q", hostPath)
	}
	return RootFS + hostPath[1:]
}

func (k *KubeBoot) String() string {
	return DebugString(k)
}
