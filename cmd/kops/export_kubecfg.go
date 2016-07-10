package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/kubecfg"
	"os"
	"path"
)

type ExportKubecfgCommand struct {
	tmpdir   string
	keyStore fi.CAStore
}

var exportKubecfgCommand ExportKubecfgCommand

func init() {
	cmd := &cobra.Command{
		Use:   "kubecfg",
		Short: "Generate a kubecfg file for a cluster",
		Long:  `Creates a kubecfg file for a cluster, based on the state`,
		Run: func(cmd *cobra.Command, args []string) {
			err := exportKubecfgCommand.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	exportCmd.AddCommand(cmd)
}

func (c *ExportKubecfgCommand) Run() error {
	clusterRegistry, cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	clusterName := cluster.Name

	master := cluster.Spec.MasterPublicName
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

	c.keyStore = clusterRegistry.KeyStore(clusterName)

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

func (c *ExportKubecfgCommand) copyCertificate(id string) (string, error) {
	p := path.Join(c.tmpdir, id+".crt")
	cert, err := c.keyStore.Cert(id)
	if err != nil {
		return "", fmt.Errorf("error fetching certificate %q: %v", id, err)
	}

	_, err = writeFile(p, cert)
	if err != nil {
		return "", fmt.Errorf("error writing certificate %q: %v", id, err)
	}

	return p, nil
}

func (c *ExportKubecfgCommand) copyPrivateKey(id string) (string, error) {
	p := path.Join(c.tmpdir, id+".key")
	cert, err := c.keyStore.PrivateKey(id)
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
