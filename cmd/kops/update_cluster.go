package main

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/kutil"
	"strings"
)

type UpdateClusterCmd struct {
	Yes    bool
	Target string
	Models string
	OutDir string
}

var updateCluster UpdateClusterCmd

func init() {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Update cluster",
		Long:  `Updates a k8s cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := updateCluster.Run(args)
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	updateCmd.AddCommand(cmd)

	cmd.Flags().BoolVar(&updateCluster.Yes, "yes", false, "Actually create cloud resources")
	cmd.Flags().StringVar(&updateCluster.Target, "target", "direct", "Target - direct, terraform")
	cmd.Flags().StringVar(&updateCluster.Models, "model", "config,proto,cloudup", "Models to apply (separate multiple models with commas)")
	cmd.Flags().StringVar(&updateCluster.OutDir, "out", "", "Path to write any local output")
}

func (c *UpdateClusterCmd) Run(args []string) error {
	err := rootCommand.ProcessArgs(args)
	if err != nil {
		return err
	}

	isDryrun := false
	// direct requires --yes (others do not, because they don't do anything!)
	if c.Target == cloudup.TargetDirect {
		if !c.Yes {
			isDryrun = true
			c.Target = cloudup.TargetDryRun
		}
	}
	if c.Target == cloudup.TargetDryRun {
		isDryrun = true
		c.Target = cloudup.TargetDryRun
	}

	if c.OutDir == "" {
		c.OutDir = "out"
	}

	clusterRegistry, cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	fullCluster, err := clusterRegistry.ReadCompletedConfig(cluster.Name)
	if err != nil {
		return err
	}

	instanceGroupRegistry, err := rootCommand.InstanceGroupRegistry()
	if err != nil {
		return err
	}

	fullInstanceGroups, err := instanceGroupRegistry.ReadAll()
	if err != nil {
		return err
	}

	strict := false
	err = api.DeepValidate(cluster, fullInstanceGroups, strict)
	if err != nil {
		return err
	}

	applyCmd := &cloudup.ApplyClusterCmd{
		Cluster:         fullCluster,
		InstanceGroups:  fullInstanceGroups,
		Models:          strings.Split(c.Models, ","),
		ClusterRegistry: clusterRegistry,
		Target:          c.Target,
		OutDir:          c.OutDir,
		DryRun:          isDryrun,
	}
	err = applyCmd.Run()
	if err != nil {
		return err
	}

	if isDryrun {
		fmt.Printf("Must specify --yes to apply changes\n")
		return nil
	}

	// TODO: Only if not yet set?
	if !isDryrun {
		keyStore := clusterRegistry.KeyStore(cluster.Name)

		kubecfgCert, err := keyStore.FindCert("kubecfg")
		if err != nil {
			// This is only a convenience; don't error because of it
			glog.Warningf("Ignoring error trying to fetch kubecfg cert - won't export kubecfg: %v", err)
			kubecfgCert = nil
		}
		if kubecfgCert != nil {
			glog.Infof("Exporting kubecfg for cluster")
			x := &kutil.CreateKubecfg{
				ClusterName:      cluster.Name,
				KeyStore:         keyStore,
				MasterPublicName: cluster.Spec.MasterPublicName,
			}
			defer x.Close()

			err = x.WriteKubecfg()
			if err != nil {
				return err
			}
		} else {
			glog.Infof("kubecfg cert not found; won't export kubecfg")
		}
	}

	return nil
}
