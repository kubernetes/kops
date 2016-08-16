package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"os"
)

type UpgradeClusterCmd struct {
	Yes bool

	NewClusterName string
}

var upgradeCluster UpgradeClusterCmd

func init() {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Upgrade cluster",
		Long:  `Upgrades a k8s cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := upgradeCluster.Run()
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().BoolVar(&upgradeCluster.Yes, "yes", false, "Apply update")

	upgradeCmd.AddCommand(cmd)
}

type upgradeAction struct {
	Item     string
	Property string
	Old      string
	New      string

	apply func()
}

func (c *UpgradeClusterCmd) Run() error {
	clusterRegistry, cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	instanceGroupRegistry, err := rootCommand.InstanceGroupRegistry()
	if err != nil {
		return err
	}

	instanceGroups, err := instanceGroupRegistry.ReadAll()

	if cluster.Annotations[api.AnnotationNameManagement] == api.AnnotationValueManagementImported {
		return fmt.Errorf("upgrade is not for use with imported clusters (did you mean `kops toolbox convert-imported`?)")
	}

	latestKubernetesVersion, err := api.FindLatestKubernetesVersion()
	if err != nil {
		return err
	}

	var actions []*upgradeAction
	if cluster.Spec.KubernetesVersion != latestKubernetesVersion {
		actions = append(actions, &upgradeAction{
			Item:     "Cluster",
			Property: "KubernetesVersion",
			Old:      cluster.Spec.KubernetesVersion,
			New:      latestKubernetesVersion,
			apply: func() {
				cluster.Spec.KubernetesVersion = latestKubernetesVersion
			},
		})
	}

	if len(actions) == 0 {
		// TODO: Allow --force option to force even if not needed?
		// Note stderr - we try not to print to stdout if no update is needed
		fmt.Fprintf(os.Stderr, "\nNo upgrade required\n")
		return nil
	}

	{
		t := &Table{}
		t.AddColumn("ITEM", func(a *upgradeAction) string {
			return a.Item
		})
		t.AddColumn("PROPERTY", func(a *upgradeAction) string {
			return a.Property
		})
		t.AddColumn("OLD", func(a *upgradeAction) string {
			return a.Old
		})
		t.AddColumn("NEW", func(a *upgradeAction) string {
			return a.New
		})

		err := t.Render(actions, os.Stdout, "ITEM", "PROPERTY", "OLD", "NEW")
		if err != nil {
			return err
		}
	}

	if !c.Yes {
		fmt.Printf("\nMust specify --yes to perform upgrade\n")
		return nil
	} else {
		for _, action := range actions {
			action.apply()
		}

		// TODO: DRY this chunk
		err = cluster.PerformAssignments()
		if err != nil {
			return fmt.Errorf("error populating configuration: %v", err)
		}

		fullCluster, err := cloudup.PopulateClusterSpec(cluster, clusterRegistry)
		if err != nil {
			return err
		}

		err = api.DeepValidate(fullCluster, instanceGroups, true)
		if err != nil {
			return err
		}

		// Note we perform as much validation as we can, before writing a bad config
		err = clusterRegistry.Update(cluster)
		if err != nil {
			return err
		}

		err = clusterRegistry.WriteCompletedConfig(fullCluster)
		if err != nil {
			return fmt.Errorf("error writing completed cluster spec: %v", err)
		}

		fmt.Printf("\nUpdates applied to configuration.\n")

		// TODO: automate this step
		fmt.Printf("You can now apply these changes, using `kops update cluster %s`\n", cluster.Name)
	}

	return nil
}
