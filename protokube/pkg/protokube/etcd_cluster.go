package protokube

import (
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
	"path"
	"strings"
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

func (k *KubeBoot) BuildEtcdClusters(modelDir string) ([]*EtcdCluster, error) {
	var clusters []*EtcdCluster

	for _, spec := range k.EtcdClusters {
		modelTemplatePath := path.Join(modelDir, spec.ClusterKey+".config")
		modelTemplate, err := ioutil.ReadFile(modelTemplatePath)
		if err != nil {
			return nil, fmt.Errorf("error reading model template %q: %v", modelTemplatePath, err)
		}

		cluster := &EtcdCluster{}
		cluster.Spec = spec

		model, err := ExecuteTemplate("model-etcd-"+spec.ClusterKey, string(modelTemplate), cluster)
		if err != nil {
			return nil, fmt.Errorf("error executing etcd model template %q: %v", modelTemplatePath, err)
		}

		err = yaml.Unmarshal([]byte(model), cluster)
		if err != nil {
			return nil, fmt.Errorf("error parsing etcd model template %q: %v", modelTemplatePath, err)
		}

		clusters = append(clusters, cluster)
	}

	return clusters, nil
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

	err := touchFile(k.PathFor(c.LogFile))
	if err != nil {
		return fmt.Errorf("error touching log-file %q: %v", c.LogFile, err)
	}

	if c.ClusterToken == "" {
		c.ClusterToken = "etcd-cluster-token-" + name
	}

	for _, nodeName := range c.Spec.NodeNames {
		name := name + "-" + nodeName
		fqdn := k.BuildInternalDNSName(name)

		node := &EtcdNode{
			Name:         name,
			InternalName: fqdn,
		}
		c.Nodes = append(c.Nodes, node)

		if nodeName == c.Spec.NodeName {
			c.Me = node

			err := k.MapInternalDNSName(fqdn)
			if err != nil {
				return fmt.Errorf("error mapping internal dns name for %q: %v", name, err)
			}
		}
	}

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
	err = ioutil.WriteFile(k.PathFor(manifestPath), []byte(manifest), 0644)
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
