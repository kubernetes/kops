package main

import (
	"fmt"

	"encoding/json"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
	"k8s.io/kube-deploy/upup/pkg/kutil"
	"strings"
)

// AddonsCmd represents the addons command
type AddonsCmd struct {
	//ClusterName  string

	cobraCommand *cobra.Command
}

var addonsCmd = AddonsCmd{
	cobraCommand: &cobra.Command{
		Use:   "addons",
		Short: "manage cluster addons",
		Long:  `manage cluster addons`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Usage: addons get")
		},
	},
}

func init() {
	cmd := addonsCmd.cobraCommand
	rootCommand.cobraCommand.AddCommand(cmd)

	//cmd.PersistentFlags().StringVar(&addonsCmd.ClusterName, "name", "", "cluster name")
}

type kubectlConfig struct {
	Kind       string                    `json:"kind`
	ApiVersion string                    `json:"apiVersion`
	Clusters   []*kubectlClusterWithName `json:"clusters`
}

type kubectlClusterWithName struct {
	Name    string         `json:"name`
	Cluster kubectlCluster `json:"cluster`
}
type kubectlCluster struct {
	Server string `json:"server`
}

func (c *AddonsCmd) buildClusterAddons() (*kutil.ClusterAddons, error) {
	//if c.ClusterName == "" {
	//	return fmt.Errorf("--name is required")
	//}

	kubectl := &kutil.Kubectl{}
	//context, err := kubectl.GetCurrentContext()
	//if err != nil {
	//	return nil, fmt.Errorf("error getting current context from kubectl: %v", err)
	//}
	//glog.V(4).Infof("context = %q", context)

	configString, err := kubectl.GetConfig(true, "json")
	if err != nil {
		return nil, fmt.Errorf("error getting current config from kubectl: %v", err)
	}
	glog.V(8).Infof("config = %q", configString)

	config := &kubectlConfig{}
	err = json.Unmarshal([]byte(configString), config)
	if err != nil {
		return nil, fmt.Errorf("cannot parse current config from kubectl: %v", err)
	}

	if len(config.Clusters) != 1 {
		return nil, fmt.Errorf("expected exactly one cluster in kubectl config, found %d", len(config.Clusters))
	}

	namedCluster := config.Clusters[0]
	glog.V(4).Infof("using cluster name %q", namedCluster.Name)
	server := namedCluster.Cluster.Server
	server = strings.TrimSpace(server)
	if server == "" {
		return nil, fmt.Errorf("server was not set in kubectl config")
	}

	k := &kutil.ClusterAddons{
		APIEndpoint: server,
	}

	privateKeyFile := utils.ExpandPath("~/.ssh/id_rsa")
	err = kutil.AddSSHIdentity(&k.SSHConfig, privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("error adding SSH private key %q: %v", err)
	}

	return k, nil
}
