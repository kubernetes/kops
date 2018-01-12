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
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/resource"
	"os"
	"path"
	"strings"
	"time"
)

type EtcdClusterSpec struct {
	ClusterKey string `json:"clusterKey,omitempty"`

	NodeName  string   `json:"nodeName,omitempty"`
	NodeNames []string `json:"nodeNames,omitempty"`
}

func (e *EtcdClusterSpec) String() string {
	return DebugString(e)
}

type EtcdCluster struct {
	PeerPort     int
	ClientPort   int
	LogFile      string
	DataDirName  string
	ClusterName  string
	ClusterToken string
	Me           *EtcdNode
	Nodes        []*EtcdNode
	PodName      string
	CPURequest   resource.Quantity

	Spec *EtcdClusterSpec

	VolumeMountPath string
}

func (e *EtcdCluster) String() string {
	return DebugString(e)
}

type EtcdNode struct {
	Name         string
	InternalName string
}

func (e *EtcdNode) String() string {
	return DebugString(e)
}

type EtcdController struct {
	kubeBoot *KubeBoot

	volume     *Volume
	volumeSpec *EtcdClusterSpec
	cluster    *EtcdCluster
}

func newEtcdController(kubeBoot *KubeBoot, v *Volume, spec *EtcdClusterSpec) (*EtcdController, error) {
	k := &EtcdController{
		kubeBoot: kubeBoot,
	}

	cluster := &EtcdCluster{}
	cluster.Spec = spec
	cluster.VolumeMountPath = v.Mountpoint

	cluster.ClusterName = "etcd-" + spec.ClusterKey
	cluster.DataDirName = "data-" + spec.ClusterKey
	cluster.PodName = "etcd-server-" + spec.ClusterKey
	cluster.CPURequest = resource.MustParse("100m")
	cluster.ClientPort = 4001
	cluster.PeerPort = 2380

	// We used to build this through text files ... it turns out to just be more complicated than code!
	switch spec.ClusterKey {
	case "main":
		cluster.ClusterName = "etcd"
		cluster.DataDirName = "data"
		cluster.PodName = "etcd-server"
		cluster.CPURequest = resource.MustParse("200m")

	case "events":
		cluster.ClientPort = 4002
		cluster.PeerPort = 2381

	default:
		return nil, fmt.Errorf("unknown Etcd ClusterKey %q", spec.ClusterKey)

	}

	k.cluster = cluster

	return k, nil
}

func (k *EtcdController) RunSyncLoop() {
	for {
		err := k.syncOnce()
		if err != nil {
			glog.Warningf("error during attempt to bootstrap (will sleep and retry): %v", err)
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

	err := touchFile(PathFor(c.LogFile))
	if err != nil {
		return fmt.Errorf("error touching log-file %q: %v", c.LogFile, err)
	}

	if c.ClusterToken == "" {
		c.ClusterToken = "etcd-cluster-token-" + name
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

			err := k.CreateInternalDNSNameRecord(fqdn)
			if err != nil {
				return fmt.Errorf("error mapping internal dns name for %q: %v", name, err)
			}
		}
	}
	c.Nodes = nodes

	if c.Me == nil {
		return fmt.Errorf("my node name %s not found in cluster %v", c.Spec.NodeName, strings.Join(c.Spec.NodeNames, ","))
	}

	pod := BuildEtcdManifest(c)
	manifest, err := ToVersionedYaml(pod)
	if err != nil {
		return fmt.Errorf("error marshalling pod to yaml: %v", err)
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
		existingManifest, err := ioutil.ReadFile(PathFor(manifestTarget))
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("error reading manifest file %q: %v", manifestTarget, err)
			}
		} else if bytes.Equal(existingManifest, manifest) {
			writeManifest = false
		} else {
			glog.Infof("Need to update manifest file: %q", manifestTarget)
		}
	}

	createSymlink := true
	{
		// See if the symlink is correct
		stat, err := os.Lstat(PathFor(manifestSource))
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("error reading manifest symlink %q: %v", manifestSource, err)
			}
		} else if (stat.Mode() & os.ModeSymlink) != 0 {
			// It's a symlink, make sure the target matches
			target, err := os.Readlink(PathFor(manifestSource))
			if err != nil {
				return fmt.Errorf("error reading manifest symlink %q: %v", manifestSource, err)
			}

			if target == manifestTarget {
				createSymlink = false
			} else {
				glog.Infof("Need to update manifest symlink (wrong target %q): %q", target, manifestSource)
			}
		} else {
			glog.Infof("Need to update manifest symlink (not a symlink): %q", manifestSource)
		}
	}

	if createSymlink || writeManifest {
		err = os.Remove(PathFor(manifestSource))
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error removing etcd manifest symlink (for strict creation) %q: %v", manifestSource, err)
		}

		err = os.MkdirAll(PathFor(manifestTargetDir), 0755)
		if err != nil {
			return fmt.Errorf("error creating directories for etcd manifest %q: %v", manifestTargetDir, err)
		}

		err = ioutil.WriteFile(PathFor(manifestTarget), manifest, 0644)
		if err != nil {
			return fmt.Errorf("error writing etcd manifest %q: %v", manifestTarget, err)
		}

		// Note: no PathFor on the target, because it's a symlink and we want it to evaluate on the host
		err = os.Symlink(manifestTarget, PathFor(manifestSource))
		if err != nil {
			return fmt.Errorf("error creating etcd manifest symlink %q -> %q: %v", manifestSource, manifestTarget, err)
		}

		glog.Infof("Updated etcd manifest: %s", manifestSource)
	}

	return nil
}

func touchFile(p string) error {
	_, err := os.Lstat(p)
	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("error getting state of file %q: %v", p, err)
	}

	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("error touching file %q: %v", p, err)
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("error closing touched file %q: %v", p, err)
	}
	return nil
}
