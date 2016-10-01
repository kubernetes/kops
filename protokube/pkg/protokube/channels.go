package protokube

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/kops/channels/pkg/channels"
	"k8s.io/kops/util/pkg/vfs"
	"net/url"
	"os"
	"os/exec"
	"path"
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

	// We copy the channel to a temp file because it is likely e.g. an s3 URL, which kubectl can't read

	location, err := url.Parse(channel)
	if err != nil {
		return fmt.Errorf("error parsing channel location: %v", err)
	}
	data, err := vfs.Context.ReadFile(channel)
	if err != nil {
		return fmt.Errorf("error reading channel: %v", err)
	}

	addons, err := channels.ParseAddons(location, data)
	if err != nil {
		return fmt.Errorf("error parsing adddons: %v", err)
	}
	all, err := addons.All()
	if err != nil {
		return fmt.Errorf("error processing adddons: %v", err)
	}

	tmpDir, err := ioutil.TempDir("", "channel")
	if err != nil {
		return fmt.Errorf("error creating temp dir: %v", err)
	}

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			glog.Warningf("error deleting temp dir: %v", err)
		}
	}()

	localChannelFile := path.Join(tmpDir, "channel.yaml")
	if err := ioutil.WriteFile(localChannelFile, data, 0600); err != nil {
		return fmt.Errorf("error writing temp file: %v", err)
	}

	for _, addon := range all {
		if addon.Spec.Manifest == nil {
			continue
		}
		manifest := *addon.Spec.Manifest
		manifestURL, err := url.Parse(manifest)
		if err != nil {
			return fmt.Errorf("error parsing manifest location: %v", manifest)
		}
		if manifestURL.IsAbs() {
			// Hopefully http or https!
			continue
		}

		dest := path.Join(tmpDir, manifest)
		src := location.ResolveReference(manifestURL)

		b, err := vfs.Context.ReadFile(src.String())
		if err != nil {
			return fmt.Errorf("error reading source manifest %q: %v", src, err)
		}

		parent := path.Dir(dest)
		if err := os.MkdirAll(parent, 0700); err != nil {
			return fmt.Errorf("error creating directories %q: %v", parent, err)
		}

		if err := ioutil.WriteFile(dest, b, 0600); err != nil {
			return fmt.Errorf("error copying channel to temp file: %v", err)
		}
	}

	out, err := execChannels("apply", "channel", localChannelFile, "--v=4", "--yes")
	glog.V(4).Infof("apply channel output was: %v", out)
	return err
}
