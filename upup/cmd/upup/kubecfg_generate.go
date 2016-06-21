package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup"
	"k8s.io/kube-deploy/upup/pkg/kubecfg"
	"os"
	"path"
)

type KubecfgGenerateCommand struct {
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
}

func (c *KubecfgGenerateCommand) Run() error {
	stateStore, err := rootCommand.StateStore()
	if err != nil {
		return fmt.Errorf("error state store: %v", err)
	}

	config := &cloudup.CloudConfig{}
	err = stateStore.ReadConfig(config)
	if err != nil {
		return fmt.Errorf("error reading configuration: %v", err)
	}

	clusterName := config.ClusterName
	if clusterName == "" {
		return fmt.Errorf("ClusterName must be set in config")
	}

	master := config.MasterPublicName
	if master == "" {
		master = "api." + clusterName
	}

	//cloudProvider := config.CloudProvider
	//if cloudProvider == "" {
	//	return fmt.Errorf("cloud must be specified")
	//}

	c.tmpdir, err = ioutil.TempDir("", "k8s")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(c.tmpdir)

	b := &kubecfg.KubeconfigBuilder{}
	b.Init()

	b.Context = clusterName
	//switch cloudProvider {
	//case "aws":
	//	b.Context = "aws_" + clusterName
	//
	//case "gce":
	//	if config.Project == "" {
	//		return fmt.Errorf("Project must be configured (for GCE)")
	//	}
	//	b.Context = config.Project + "_" + clusterName
	//
	//default:
	//	return fmt.Errorf("Unknown cloud provider %q", cloudProvider)
	//}

	c.caStore, err = rootCommand.CA()
	if err != nil {
		return err
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

	b.KubeMasterIP = master

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
