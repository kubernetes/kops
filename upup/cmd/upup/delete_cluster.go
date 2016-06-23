package main

import (
	"fmt"

	"bytes"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/kutil"
	"os"
	"reflect"
	"text/tabwriter"
)

type DeleteClusterCmd struct {
	ClusterName string
	Yes         bool
	Region      string
}

var deleteCluster DeleteClusterCmd

func init() {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Delete cluster",
		Long:  `Deletes a k8s cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := deleteCluster.Run()
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	deleteCmd.AddCommand(cmd)

	cmd.Flags().BoolVar(&deleteCluster.Yes, "yes", false, "Delete without confirmation")

	cmd.Flags().StringVar(&deleteCluster.ClusterName, "name", "", "cluster name")
	cmd.Flags().StringVar(&deleteCluster.Region, "region", "", "region")
}

type getter func(o interface{}) interface{}

func (c *DeleteClusterCmd) Run() error {
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

	d := &kutil.DeleteCluster{}

	d.ClusterName = c.ClusterName
	d.Region = c.Region
	d.Cloud = cloud

	glog.Infof("TODO: S3 bucket removal")

	resources, err := d.ListResources()
	if err != nil {
		return err
	}

	columns := []string{"TYPE", "ID", "NAME"}
	fields := []string{"Type", "ID", "Name"}

	var b bytes.Buffer
	w := new(tabwriter.Writer)

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', tabwriter.StripEscape)

	writeHeader := true
	if writeHeader {
		for i, c := range columns {
			if i != 0 {
				b.WriteByte('\t')
			}
			b.WriteByte(tabwriter.Escape)
			b.WriteString(c)
			b.WriteByte(tabwriter.Escape)
		}
		b.WriteByte('\n')

		_, err := w.Write(b.Bytes())
		if err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
		b.Reset()
	}

	for _, t := range resources {
		for i := range columns {
			if i != 0 {
				b.WriteByte('\t')
			}

			v := reflect.ValueOf(t)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			fv := v.FieldByName(fields[i])

			s := fi.ValueAsString(fv)

			b.WriteByte(tabwriter.Escape)
			b.WriteString(s)
			b.WriteByte(tabwriter.Escape)
		}
		b.WriteByte('\n')

		_, err := w.Write(b.Bytes())
		if err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
		b.Reset()
	}
	w.Flush()

	if len(resources) == 0 {
		fmt.Printf("Nothing to delete\n")
		return nil
	}

	if !c.Yes {
		return fmt.Errorf("Must specify --yes to delete")
	}

	return d.DeleteResources(resources)
}
