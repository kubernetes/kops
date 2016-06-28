package protokube

import (
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

	manifestPath := "/etc/kubernetes/manifests/" + name + ".manifest"
	err = ioutil.WriteFile(PathFor(manifestPath), []byte(manifest), 0644)
	if err != nil {
		return fmt.Errorf("error writing etcd manifest %q: %v", manifestPath, err)
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
