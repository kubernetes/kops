package channels

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/kops/util/pkg/vfs"
	"os"
	"os/exec"
	"path"
	"strings"
)

// Apply calls kubectl apply to apply the manifest.
// We will likely in future change this to create things directly (or more likely embed this logic into kubectl itself)
func Apply(manifest string) error {
	// We copy the manifest to a temp file because it is likely e.g. an s3 URL, which kubectl can't read
	data, err := vfs.Context.ReadFile(manifest)
	if err != nil {
		return fmt.Errorf("error reading manifest: %v", err)
	}

	tmpDir, err := ioutil.TempDir("", "channel")
	if err != nil {
		return fmt.Errorf("error creating temp dir: %v", err)
	}

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			glog.Warningf("error deleting temp dir %q: %v", tmpDir, err)
		}
	}()

	localManifestFile := path.Join(tmpDir, "manifest.yaml")
	if err := ioutil.WriteFile(localManifestFile, data, 0600); err != nil {
		return fmt.Errorf("error writing temp file: %v", err)
	}

	_, err = execKubectl("apply", "-f", localManifestFile)
	return err
}

func execKubectl(args ...string) (string, error) {
	kubectlPath := "kubectl" // Assume in PATH
	cmd := exec.Command(kubectlPath, args...)
	env := os.Environ()
	cmd.Env = env

	human := strings.Join(cmd.Args, " ")
	glog.V(2).Infof("Running command: %s", human)
	output, err := cmd.CombinedOutput()
	if err != nil {
		glog.Infof("error running %s", human)
		glog.Info(string(output))
		return string(output), fmt.Errorf("error running kubectl")
	}

	return string(output), err
}
