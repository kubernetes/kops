package main

import (
	"fmt"

	"bytes"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"os"
	"path"
	"text/tabwriter"
)

type GetSecretsCommand struct {
	StateDir string
}

var getSecretsCommand GetSecretsCommand

func init() {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get secrets",
		Long:  `Get secrets.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := getSecretsCommand.Run()
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	secretsCmd.AddCommand(cmd)

	cmd.Flags().StringVarP(&getSecretsCommand.StateDir, "state", "", "", "Directory in which to store state")
}

type SecretInfo struct {
	Id   string
	Type string
}

func (c *GetSecretsCommand) Run() error {
	if c.StateDir == "" {
		return fmt.Errorf("state dir is required")
	}

	var infos []*SecretInfo
	{
		caStore, err := fi.NewFilesystemCAStore(path.Join(c.StateDir, "pki"))
		if err != nil {
			return fmt.Errorf("error building CA store: %v", err)
		}
		ids, err := caStore.List()
		if err != nil {
			return fmt.Errorf("error listing CA store items %v", err)
		}

		for _, id := range ids {
			info := &SecretInfo{
				Id:   id,
				Type: "keypair",
			}
			infos = append(infos, info)
		}
	}

	{
		secretStore, err := fi.NewFilesystemSecretStore(path.Join(c.StateDir, "secrets"))
		if err != nil {
			return fmt.Errorf("error building secret store: %v", err)
		}
		ids, err := secretStore.ListSecrets()
		if err != nil {
			return fmt.Errorf("error listing secrets %v", err)
		}

		for _, id := range ids {
			info := &SecretInfo{
				Id:   id,
				Type: "secret",
			}
			infos = append(infos, info)
		}
	}

	var b bytes.Buffer
	w := new(tabwriter.Writer)

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', tabwriter.StripEscape)
	for _, info := range infos {
		b.WriteByte(tabwriter.Escape)
		b.WriteString(info.Type)
		b.WriteByte(tabwriter.Escape)
		b.WriteByte('\t')
		b.WriteByte(tabwriter.Escape)
		b.WriteString(info.Id)
		b.WriteByte(tabwriter.Escape)
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
