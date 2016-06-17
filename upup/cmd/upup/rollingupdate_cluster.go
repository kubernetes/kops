package main

import (
	"fmt"

	"bytes"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/kutil"
	"os"
	"text/tabwriter"
)

type RollingUpdateClusterCmd struct {
	ClusterName string
	Yes         bool
	Region      string

	cobraCommand *cobra.Command
}

var rollingupdateCluster = RollingUpdateClusterCmd{
	cobraCommand: &cobra.Command{
		Use:   "cluster",
		Short: "rolling-update cluster",
		Long:  `rolling-updates a k8s cluster.`,
	},
}

func init() {
	cmd := rollingupdateCluster.cobraCommand
	rollingUpdateCommand.cobraCommand.AddCommand(cmd)

	cmd.Flags().BoolVar(&rollingupdateCluster.Yes, "yes", false, "Rollingupdate without confirmation")

	cmd.Flags().StringVar(&rollingupdateCluster.ClusterName, "name", "", "cluster name")
	cmd.Flags().StringVar(&rollingupdateCluster.Region, "region", "", "region")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := rollingupdateCluster.Run()
		if err != nil {
			glog.Exitf("%v", err)
		}
	}
}

func (c *RollingUpdateClusterCmd) Run() error {
	if c.Region == "" {
		return fmt.Errorf("--region is required")
	}
	if c.ClusterName == "" {
		return fmt.Errorf("--name is required")
	}

	tags := map[string]string{"KubernetesCluster": c.ClusterName}
	cloud, err := awsup.NewAWSCloud(c.Region, tags)
	if err != nil {
		return fmt.Errorf("error initializing AWS client: %v", err)
	}

	d := &kutil.RollingUpdateCluster{}

	d.ClusterName = c.ClusterName
	d.Region = c.Region
	d.Cloud = cloud

	nodesets, err := d.ListNodesets()
	if err != nil {
		return err
	}

	err = c.printNodesets(nodesets)
	if err != nil {
		return err
	}

	if !c.Yes {
		return fmt.Errorf("Must specify --yes to rolling-update")
	}

	return d.RollingUpdateNodesets(nodesets)
}

func (c *RollingUpdateClusterCmd) printNodesets(nodesets map[string]*kutil.Nodeset) error {
	w := new(tabwriter.Writer)
	var b bytes.Buffer

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', tabwriter.StripEscape)
	for _, n := range nodesets {
		b.WriteByte(tabwriter.Escape)
		b.WriteString(n.Name)
		b.WriteByte(tabwriter.Escape)
		b.WriteByte('\t')
		b.WriteByte(tabwriter.Escape)
		b.WriteString(n.Status)
		b.WriteByte(tabwriter.Escape)
		b.WriteByte('\t')
		b.WriteByte(tabwriter.Escape)
		b.WriteString(fmt.Sprintf("%d", len(n.NeedUpdate)))
		b.WriteByte(tabwriter.Escape)
		b.WriteByte('\t')
		b.WriteByte(tabwriter.Escape)
		b.WriteString(fmt.Sprintf("%d", len(n.Ready)))
		b.WriteByte(tabwriter.Escape)
		b.WriteByte('\n')

		_, err := w.Write(b.Bytes())
		if err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
		b.Reset()
	}

	return w.Flush()
}
