package main

import (
	"fmt"

	"bytes"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/upup/pkg/api"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"os"
	"reflect"
	"text/tabwriter"
)

type GetClustersCmd struct {
}

var getClustersCmd GetClustersCmd

func init() {
	cmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"clusters"},
		Short:   "get clusters",
		Long:    `List or get clusters.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := getClustersCmd.Run()
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	getCmd.AddCommand(cmd)
}

func (c *GetClustersCmd) Run() error {
	clusterNames, err := rootCommand.ListClusters()
	if err != nil {
		return err
	}

	columns := []string{}
	fields := []func(*api.Cluster) string{}

	columns = append(columns, "NAME")
	fields = append(fields, func(c *api.Cluster) string {
		return c.Name
	})

	var clusters []*api.Cluster

	for _, clusterName := range clusterNames {
		stateStore, err := rootCommand.StateStoreForCluster(clusterName)
		if err != nil {
			return err
		}

		// TODO: Faster if we don't read groups...
		// We probably can just have a comand which directly reads all cluster config files
		cluster, _, err := api.ReadConfig(stateStore)
		clusters = append(clusters, cluster)
	}
	if len(clusters) == 0 {
		return nil
	}
	return WriteTable(clusters, columns, fields)
}

func WriteTable(items interface{}, columns []string, fields interface{}) error {
	itemsValue := reflect.ValueOf(items)
	if itemsValue.Kind() != reflect.Slice {
		glog.Fatal("unexpected kind for items in WriteTable: ", itemsValue.Kind())
	}
	fieldsValue := reflect.ValueOf(fields)
	if fieldsValue.Kind() != reflect.Slice {
		glog.Fatal("unexpected kind for fields in WriteTable: ", fieldsValue.Kind())
	}

	length := itemsValue.Len()

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

	for i := 0; i < length; i++ {
		item := itemsValue.Index(i)

		for j := range columns {
			if j != 0 {
				b.WriteByte('\t')
			}

			fieldFunc := fieldsValue.Index(j)
			var args []reflect.Value
			args = append(args, item)
			fvs := fieldFunc.Call(args)
			fv := fvs[0]

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

	return nil
}
