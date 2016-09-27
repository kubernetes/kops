package channels

import (
	"fmt"
	"github.com/golang/glog"
	"os"
	"os/exec"
	"strings"
)

// Apply calls kubectl apply to apply the manifest.
// We will likely in future change this to create things directly (or more likely embed this logic into kubectl itself)
func Apply(manifest string) error {
	_, err := execKubectl("apply", "-f", manifest)
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
