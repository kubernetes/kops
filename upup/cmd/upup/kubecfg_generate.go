package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/kubecfg"
	"os"
	"path"
)

type KubecfgGenerateCommand struct {
	StateDir      string
	ClusterName   string
	CloudProvider string
	Project       string
	Master        string

	tmpdir  string
	caStore fi.CAStore
}

var kubecfgGenerateCommand KubecfgGenerateCommand

func init() {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a kubecfg file for a cluster",
		Long:  `Creates a kubecfg file for a cluster, based on the state`,
		Run: func(cmd *cobra.Command, args []string) {
			err := kubecfgGenerateCommand.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	kubecfgCmd.AddCommand(cmd)

	// TODO: We need to store this in the persistent state dir
	cmd.Flags().StringVarP(&kubecfgGenerateCommand.ClusterName, "name", "", kubecfgGenerateCommand.ClusterName, "Name for cluster")
	cmd.Flags().StringVarP(&kubecfgGenerateCommand.CloudProvider, "cloud", "", kubecfgGenerateCommand.CloudProvider, "Cloud provider to use - gce, aws")
	cmd.Flags().StringVarP(&kubecfgGenerateCommand.Project, "project", "", kubecfgGenerateCommand.Project, "Project to use (must be set on GCE)")

	cmd.Flags().StringVarP(&kubecfgGenerateCommand.Master, "master", "", kubecfgGenerateCommand.Master, "IP adddress or host of API server")

	cmd.Flags().StringVarP(&kubecfgGenerateCommand.StateDir, "state", "", "", "State directory")
}

func (c *KubecfgGenerateCommand) Run() error {
	if c.StateDir == "" {
		return fmt.Errorf("state must be specified")
	}

	if c.Master == "" {
		return fmt.Errorf("master must be specified")
	}

	if c.ClusterName == "" {
		return fmt.Errorf("name must be specified")
	}
	if c.CloudProvider == "" {
		return fmt.Errorf("cloud must be specified")
	}

	var err error
	c.tmpdir, err = ioutil.TempDir("", "k8s")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(c.tmpdir)

	b := &kubecfg.KubeconfigBuilder{}
	b.Init()

	switch c.CloudProvider {
	case "aws":
		b.Context = "aws_" + c.ClusterName

	case "gce":
		if c.Project == "" {
			return fmt.Errorf("--project must be specified (for GCE)")
		}
		b.Context = c.Project + "_" + c.ClusterName

	default:
		return fmt.Errorf("Unknown cloud provider %q", c.CloudProvider)
	}

	c.caStore, err = fi.NewFilesystemCAStore(path.Join(c.StateDir, "pki"))
	if err != nil {
		return fmt.Errorf("error building CA store: %v", err)
	}

	if b.CACert, err = c.copyCertificate(fi.CertificateId_CA); err != nil {
		return err
	}

	if b.KubecfgCert, err = c.copyCertificate("kubecfg"); err != nil {
		return err
	}

	if b.KubecfgKey, err = c.copyPrivateKey("kubecfg"); err != nil {
		return err
	}

	b.KubeMasterIP = c.Master

	err = b.CreateKubeconfig()
	if err != nil {
		return err
	}

	return nil
}

func (c *KubecfgGenerateCommand) copyCertificate(id string) (string, error) {
	p := path.Join(c.tmpdir, id+".crt")
	cert, err := c.caStore.Cert(id)
	if err != nil {
		return "", fmt.Errorf("error fetching certificate %q: %v", id, err)
	}

	_, err = writeFile(p, cert)
	if err != nil {
		return "", fmt.Errorf("error writing certificate %q: %v", id, err)
	}

	return p, nil
}

func (c *KubecfgGenerateCommand) copyPrivateKey(id string) (string, error) {
	p := path.Join(c.tmpdir, id+".key")
	cert, err := c.caStore.PrivateKey(id)
	if err != nil {
		return "", fmt.Errorf("error fetching private key %q: %v", id, err)
	}

	_, err = writeFile(p, cert)
	if err != nil {
		return "", fmt.Errorf("error writing private key %q: %v", id, err)
	}

	return p, nil
}

func writeFile(dst string, src io.WriterTo) (int64, error) {
	f, err := os.Create(dst)
	if err != nil {
		return 0, fmt.Errorf("error creating file %q: %v", dst, err)
	}
	defer fi.SafeClose(f)
	return src.WriteTo(f)
}
