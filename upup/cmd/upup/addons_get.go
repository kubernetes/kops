package main

import (
	"fmt"

	"bytes"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/upup/pkg/kutil"
	"os"
	"text/tabwriter"
)

type AddonsGetCmd struct {
	ClusterName string

	cobraCommand *cobra.Command
}

var addonsGetCmd = AddonsGetCmd{
	cobraCommand: &cobra.Command{
		Use:   "get",
		Short: "Display one or many addons",
		Long:  `Query a cluster, and list the addons.`,
	},
}

func init() {
	cmd := addonsGetCmd.cobraCommand
	addonsCmd.cobraCommand.AddCommand(cmd)

	cmd.Flags().StringVar(&addonsGetCmd.ClusterName, "name", "", "cluster name")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := addonsGetCmd.Run()
		if err != nil {
			glog.Exitf("%v", err)
		}
	}
}

func (c *AddonsGetCmd) Run() error {
	k, err := addonsCmd.buildClusterAddons()
	if err != nil {
		return err
	}

	addons, err := k.ListAddons()
	if err != nil {
		return err
	}

	err = c.printAddons(addons)
	if err != nil {
		return err
	}

	return nil
}

func (c *AddonsGetCmd) printAddons(addons map[string]*kutil.ClusterAddon) error {
	w := new(tabwriter.Writer)
	var b bytes.Buffer

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', tabwriter.StripEscape)
	for _, n := range addons {
		b.WriteByte(tabwriter.Escape)
		b.WriteString(n.Name)
		b.WriteByte(tabwriter.Escape)
		b.WriteByte('\t')
		b.WriteByte(tabwriter.Escape)
		b.WriteString(n.Path)
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
