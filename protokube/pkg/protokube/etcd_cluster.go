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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog"
	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/protokube/pkg/etcd"
)

// EtcdCluster is the configuration for the etcd cluster
type EtcdCluster struct {
	// ClientPort is the incoming ports for client
	ClientPort int
	// ClusterName is the cluster name
	ClusterName string
	// ClusterToken is the cluster token
	ClusterToken string
	// CPURequest is the pod request for CPU
	CPURequest *resource.Quantity
	// MemoryRequest is the pod request for Memory
	MemoryRequest *resource.Quantity
	// DataDirName is the path to the data directory
	DataDirName string
	// ImageSource is the docker image to use
	ImageSource string
	// LogFile is the location of the logfile
	LogFile string
	// Me is the node that we will be in the cluster
	Me *EtcdNode
	// Nodes is a list of nodes in the cluster (including the self-node, Me)
	Nodes []*EtcdNode
	// PeerPort is the port for peers to connect
	PeerPort int
	// PodName is the name given to the pod
	PodName string
	// ProxyMode indicates we are running in proxy mode
	ProxyMode bool
	// Spec is the specification found from the volumes
	Spec *etcd.EtcdClusterSpec
	// VolumeMountPath is the mount path
	VolumeMountPath string
	// TLSAuth indicates we should enforce peer and client verification
	TLSAuth bool
	// TLSCA is the path to a client ca for etcd clients
	TLSCA string
	// TLSCert is the path to a client certificate for etcd
	TLSCert string
	// TLSKey is the path to a client private key for etcd
	TLSKey string
	// PeerCA is the path to a peer ca for etcd
	PeerCA string
	// PeerCert is the path to a peer ca for etcd
	PeerCert string
	// PeerKey is the path to a peer ca for etcd
	PeerKey string
	// ElectionTimeout is the leader election timeout
	ElectionTimeout string
	// HeartbeatInterval is the heartbeat interval
	HeartbeatInterval string
	// BackupImage is the image to use for backing up etcd
	BackupImage string
	// BackupStore is a VFS path for backing up etcd
	BackupStore string
}

// EtcdNode is a definition for the etcd node
type EtcdNode struct {
	Name         string
	InternalName string
}

// EtcdController defines the etcd controller
type EtcdController struct {
	kubeBoot *KubeBoot
	//volume     *Volume
	//volumeSpec *etcd.EtcdClusterSpec
	cluster *EtcdCluster
}

// newEtcdController creates and returns a new etcd controller
func newEtcdController(kubeBoot *KubeBoot, v *Volume, spec *etcd.EtcdClusterSpec) (*EtcdController, error) {
	k := &EtcdController{
		kubeBoot: kubeBoot,
	}

	// prepare parse variables for cpu and memory requests
	cpuRequest100 := resource.MustParse("100m")
	cpuRequest200 := resource.MustParse("200m")
	memoryRequest := resource.MustParse("100Mi")

	cluster := &EtcdCluster{
		CPURequest:        &cpuRequest100,
		MemoryRequest:     &memoryRequest,
		ClientPort:        4001,
		ClusterName:       "etcd-" + spec.ClusterKey,
		DataDirName:       "data-" + spec.ClusterKey,
		ElectionTimeout:   kubeBoot.EtcdElectionTimeout,
		HeartbeatInterval: kubeBoot.EtcdHeartbeatInterval,
		ImageSource:       kubeBoot.EtcdImageSource,
		PeerCA:            kubeBoot.PeerCA,
		PeerCert:          kubeBoot.PeerCert,
		PeerKey:           kubeBoot.PeerKey,
		PeerPort:          2380,
		PodName:           "etcd-server-" + spec.ClusterKey,
		Spec:              spec,
		TLSAuth:           kubeBoot.TLSAuth,
		TLSCA:             kubeBoot.TLSCA,
		TLSCert:           kubeBoot.TLSCert,
		TLSKey:            kubeBoot.TLSKey,
		VolumeMountPath:   v.Mountpoint,
	}

	// We used to build this through text files ... it turns out to just be more complicated than code!
	switch spec.ClusterKey {
	case "main":
		cluster.ClusterName = "etcd"
		cluster.DataDirName = "data"
		cluster.PodName = "etcd-server"
		cluster.CPURequest = &cpuRequest200

		// Because we can only specify a single EtcdBackupStore at the moment, we only backup main, not events
		cluster.BackupImage = kubeBoot.EtcdBackupImage
		cluster.BackupStore = kubeBoot.EtcdBackupStore

	case "events":
		cluster.ClientPort = 4002
		cluster.PeerPort = 2381
	default:
		return nil, fmt.Errorf("unknown etcd cluster key %q", spec.ClusterKey)
	}

	k.cluster = cluster

	return k, nil
}

// RunSyncLoop is responsible for managing the etcd sign loop
func (k *EtcdController) RunSyncLoop() {
	for {
		if err := k.syncOnce(); err != nil {
			klog.Warningf("error during attempt to bootstrap (will sleep and retry): %v", err)
		}

		time.Sleep(1 * time.Minute)
	}
}

func (k *EtcdController) syncOnce() error {
	return k.cluster.configure(k.kubeBoot)
}

func (c *EtcdCluster) configure(k *KubeBoot) error {
	name := c.ClusterName
	if !strings.HasPrefix(name, "etcd") {
		// For sanity, and to avoid collisions in directories / dns
		return fmt.Errorf("unexpected name for etcd cluster (must start with etcd): %q", name)
	}
	if c.LogFile == "" {
		c.LogFile = "/var/log/" + name + ".log"
	}

	if c.PodName == "" {
		c.PodName = c.ClusterName
	}

	err := touchFile(pathFor(c.LogFile))
	if err != nil {
		return fmt.Errorf("error touching log-file %q: %v", c.LogFile, err)
	}

	if c.ClusterToken == "" {
		c.ClusterToken = "etcd-cluster-token-" + name
	}

	// By default we use 100Mi for etcd memory
	if c.MemoryRequest == nil || c.MemoryRequest.IsZero() {
		memoryRequest, err := resource.ParseQuantity("100Mi")
		if err != nil {
			return fmt.Errorf("error parsing memory request for etcd (%s): %v", "100Mi", err)
		}
		c.MemoryRequest = &memoryRequest
	}

	// By default we use 100m for etcd cpu, unless the name is 'main', then we use 200m by default
	if c.CPURequest == nil || c.CPURequest.IsZero() {
		if c.ClusterName == "main" {
			cpuRequest, err := resource.ParseQuantity("200m")
			if err != nil {
				return fmt.Errorf("error parsing cpu request for etcd (%s): %v", "200m", err)
			}
			c.CPURequest = &cpuRequest
		} else {
			cpuRequest, err := resource.ParseQuantity("100m")
			if err != nil {
				return fmt.Errorf("error parsing cpu request for etcd (%s): %v", "100m", err)
			}
			c.CPURequest = &cpuRequest
		}
	}

	var nodes []*EtcdNode
	for _, nodeName := range c.Spec.NodeNames {
		name := name + "-" + nodeName
		fqdn := k.BuildInternalDNSName(name)

		node := &EtcdNode{
			Name:         name,
			InternalName: fqdn,
		}
		nodes = append(nodes, node)
		if nodeName == c.Spec.NodeName {
			c.Me = node
			if err = k.CreateInternalDNSNameRecord(fqdn); err != nil {
				return fmt.Errorf("error mapping internal dns name for %q: %v", name, err)
			}
		}
	}
	c.Nodes = nodes

	if c.Me == nil {
		return fmt.Errorf("my node name %s not found in cluster %v", c.Spec.NodeName, strings.Join(c.Spec.NodeNames, ","))
	}

	pod := BuildEtcdManifest(c)
	manifest, err := k8scodecs.ToVersionedYaml(pod)
	if err != nil {
		return fmt.Errorf("error marshaling pod to yaml: %v", err)
	}

	// Time to write the manifest!

	// To avoid a possible race condition where the manifest survives a reboot but the volume
	// is not mounted or not yet mounted, we use a symlink from /etc/kubernetes/manifests/<name>.manifest
	// to a file on the volume itself.  Thus kubelet cannot launch the manifest unless the volume is mounted.

	manifestSource := "/etc/kubernetes/manifests/" + name + ".manifest"
	manifestTargetDir := path.Join(c.VolumeMountPath, "k8s.io", "manifests")
	manifestTarget := path.Join(manifestTargetDir, name+".manifest")

	writeManifest := true
	{
		// See if the manifest has changed
		existingManifest, err := ioutil.ReadFile(pathFor(manifestTarget))
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("error reading manifest file %q: %v", manifestTarget, err)
			}
		} else if bytes.Equal(existingManifest, manifest) {
			writeManifest = false
		} else {
			klog.Infof("Need to update manifest file: %q", manifestTarget)
		}
	}

	createSymlink := true
	{
		// See if the symlink is correct
		stat, err := os.Lstat(pathFor(manifestSource))
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("error reading manifest symlink %q: %v", manifestSource, err)
			}
		} else if (stat.Mode() & os.ModeSymlink) != 0 {
			// It's a symlink, make sure the target matches
			target, err := os.Readlink(pathFor(manifestSource))
			if err != nil {
				return fmt.Errorf("error reading manifest symlink %q: %v", manifestSource, err)
			}

			if target == manifestTarget {
				createSymlink = false
			} else {
				klog.Infof("Need to update manifest symlink (wrong target %q): %q", target, manifestSource)
			}
		} else {
			klog.Infof("Need to update manifest symlink (not a symlink): %q", manifestSource)
		}
	}

	if createSymlink || writeManifest {
		err = os.Remove(pathFor(manifestSource))
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error removing etcd manifest symlink (for strict creation) %q: %v", manifestSource, err)
		}

		err = os.MkdirAll(pathFor(manifestTargetDir), 0755)
		if err != nil {
			return fmt.Errorf("error creating directories for etcd manifest %q: %v", manifestTargetDir, err)
		}

		err = ioutil.WriteFile(pathFor(manifestTarget), manifest, 0644)
		if err != nil {
			return fmt.Errorf("error writing etcd manifest %q: %v", manifestTarget, err)
		}

		// Note: no pathFor on the target, because it's a symlink and we want it to evaluate on the host
		err = os.Symlink(manifestTarget, pathFor(manifestSource))
		if err != nil {
			return fmt.Errorf("error creating etcd manifest symlink %q -> %q: %v", manifestSource, manifestTarget, err)
		}

		klog.Infof("Updated etcd manifest: %s", manifestSource)
	}

	return nil
}

// isTLS indicates the etcd cluster should be configured to use tls
func (c *EtcdCluster) isTLS() bool {
	return notEmpty(c.TLSCert) && notEmpty(c.TLSKey)
}

// String returns the debug string
func (c *EtcdCluster) String() string {
	return DebugString(c)
}

func (e *EtcdNode) String() string {
	return DebugString(e)
}
