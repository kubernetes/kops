package protokube

import (
	"fmt"
	"github.com/golang/glog"
	"os"
	"os/exec"
	"strings"
)

func execChannels(args ...string) (string, error) {
	kubectlPath := "channels" // Assume in PATH
	cmd := exec.Command(kubectlPath, args...)
	env := os.Environ()
	cmd.Env = env

	human := strings.Join(cmd.Args, " ")
	glog.V(2).Infof("Running command: %s", human)
	output, err := cmd.CombinedOutput()
	if err != nil {
		glog.Infof("error running %s:", human)
		glog.Info(string(output))
		return string(output), fmt.Errorf("error running channels: %v", err)
	}

	return string(output), err
}

func ApplyChannel(channel string) error {
	// We don't embed the channels code because we expect this will eventually be part of kubectl
	glog.Infof("checking channel: %q", channel)

	out, err := execChannels("apply", "channel", channel, "--v=4", "--yes")
	glog.V(4).Infof("apply channel output was: %v", out)
	return err
}
