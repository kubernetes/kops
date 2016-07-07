package protokube

import (
	"bytes"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

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

	modelTemplatePath := path.Join(kubeBoot.ModelDir, spec.ClusterKey+".config")
	modelTemplate, err := ioutil.ReadFile(modelTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("error reading model template %q: %v", modelTemplatePath, err)
	}

	cluster := &EtcdCluster{}
	cluster.Spec = spec
	cluster.VolumeMountPath = v.Mountpoint

	model, err := ExecuteTemplate("model-etcd-"+spec.ClusterKey, string(modelTemplate), cluster)
	if err != nil {
		return nil, fmt.Errorf("error executing etcd model template %q: %v", modelTemplatePath, err)
	}

	err = yaml.Unmarshal([]byte(model), cluster)
	if err != nil {
		return nil, fmt.Errorf("error parsing etcd model template %q: %v", modelTemplatePath, err)
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

	manifestTemplatePath := "templates/etcd/manifest.template"
	manifestTemplate, err := ioutil.ReadFile(manifestTemplatePath)
	if err != nil {
		return fmt.Errorf("error reading etcd manifest template %q: %v", manifestTemplatePath, err)
	}
	manifest, err := ExecuteTemplate("etcd-manifest", string(manifestTemplate), c)
	if err != nil {
		return fmt.Errorf("error executing etcd manifest template: %v", err)
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
