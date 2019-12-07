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

package protokube

import (
	"fmt"
	"net"
	"path/filepath"
	"time"

	"k8s.io/klog"
	utilexec "k8s.io/utils/exec"
	"k8s.io/utils/nsenter"
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
	// Kubernetes holds a kubernetes client
	Kubernetes *KubernetesContext
	// Master indicates we are a master node
	Master bool

	// ManageEtcd is true if we should manage etcd.
	// Deprecated in favor of etcd-manager.
	ManageEtcd bool
	// EtcdBackupImage is the image to use for backing up etcd
	EtcdBackupImage string
	// EtcdBackupStore is the VFS path to which we should backup etcd
	EtcdBackupStore string
	// Etcd container registry location.
	EtcdImageSource string
	// EtcdElectionTimeout is the leader election timeout
	EtcdElectionTimeout string
	// EtcdHeartbeatInterval is the heartbeat interval
	EtcdHeartbeatInterval string
	// TLSAuth indicates we should enforce peer and client verification
	TLSAuth bool
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

	// BootstrapMasterNodeLabels controls the initial application of node labels to our node
	// The node is found by matching NodeName
	BootstrapMasterNodeLabels bool

	// NodeName is the name of our node as it will be registered in k8s.
	// Used by BootstrapMasterNodeLabels
	NodeName string

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
			klog.Warningf("error during attempt to bootstrap (will sleep and retry): %v", err)
		}

		time.Sleep(1 * time.Minute)
	}
}

func (k *KubeBoot) syncOnce() error {
	if k.Master && k.ManageEtcd {
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
					klog.Infof("Found etcd cluster spec on volume %q: %v", v.ID, etcdSpec)
					etcdController, err := newEtcdController(k, v, etcdSpec)
					if err != nil {
						klog.Warningf("error building etcd controller: %v", err)
					} else {
						k.etcdControllers[key] = etcdController
						go etcdController.RunSyncLoop()
					}
				}
			}
		}
	} else if k.ManageEtcd {
		klog.V(4).Infof("Not in role master; won't scan for volumes")
	} else {
		klog.V(4).Infof("protokube management of etcd not enabled; won't scan for volumes")
	}

	// Ensure kubelet is running. We avoid doing this automatically so
	// that when kubelet comes up the first time, all volume mounts
	// and DNS are available, avoiding the scenario where
	// etcd/apiserver retry too many times and go into backoff.
	if err := startKubeletService(); err != nil {
		klog.Warningf("error ensuring kubelet started: %v", err)
	}

	if k.Master {
		if k.BootstrapMasterNodeLabels {
			if err := bootstrapMasterNodeLabels(k.Kubernetes, k.NodeName); err != nil {
				klog.Warningf("error bootstrapping master node labels: %v", err)
			}
		}
		if k.ApplyTaints {
			if err := applyMasterTaints(k.Kubernetes); err != nil {
				klog.Warningf("error updating master taints: %v", err)
			}
		}
		if k.InitializeRBAC {
			if err := applyRBAC(k.Kubernetes); err != nil {
				klog.Warningf("error initializing rbac: %v", err)
			}
		}
		for _, channel := range k.Channels {
			if err := applyChannel(channel); err != nil {
				klog.Warningf("error applying channel %q: %v", channel, err)
			}
		}
	}

	return nil
}

// startKubeletService is responsible for checking and if not starting the kubelet service
func startKubeletService() error {
	// TODO: Check/log status of kubelet
	// (in particular, we want to avoid kubernetes/kubernetes#40123 )
	klog.V(2).Infof("ensuring that kubelet systemd service is running")

	// We run systemctl from the hostfs so we don't need systemd in our image
	// (and we don't risk version skew)

	exec := utilexec.New()
	if Containerized {
		e, err := nsenter.NewNsenter(pathFor("/"), utilexec.New())
		if err != nil {
			return fmt.Errorf("error building nsenter executor: %v", err)
		}
		exec = e
	}

	systemctlCommand := "systemctl"

	output, err := exec.Command(systemctlCommand, "status", "--no-block", "kubelet").CombinedOutput()
	klog.V(2).Infof("'systemctl status kubelet' output:\n%s", string(output))
	if err == nil {
		klog.V(2).Infof("kubelet systemd service already running")
		return nil
	}

	klog.Infof("kubelet systemd service not running. Starting")
	output, err = exec.Command(systemctlCommand, "start", "--no-block", "kubelet").CombinedOutput()
	if err != nil {
		return fmt.Errorf("error starting kubelet: %v\nOutput: %s", err, output)
	}
	klog.V(2).Infof("'systemctl start kubelet' output:\n%s", string(output))

	return nil
}

func pathFor(hostPath string) string {
	if hostPath[0] != '/' {
		klog.Fatalf("path was not absolute: %q", hostPath)
	}
	return RootFS + hostPath[1:]
}

func pathForSymlinks(hostPath string) string {
	path := pathFor(hostPath)

	symlink, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}

	return symlink
}

func (k *KubeBoot) String() string {
	return DebugString(k)
}
