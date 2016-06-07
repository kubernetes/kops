package nodetasks

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kube-deploy/upup/pkg/fi/nodeup/local"
	"os"
	"os/exec"
)

type UpdatePackages struct {
}

var _ fi.HasDependencies = &UpdatePackages{}

func (p *UpdatePackages) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return []fi.Task{}
}

func (p *UpdatePackages) String() string {
	return fmt.Sprintf("UpdatePackages")
}

func (e *UpdatePackages) Find(c *fi.Context) (*UpdatePackages, error) {
	return nil, nil
}

func (e *UpdatePackages) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *UpdatePackages) CheckChanges(a, e, changes *Service) error {
	return nil
}

func (_ *UpdatePackages) RenderLocal(t *local.LocalTarget, a, e, changes *UpdatePackages) error {
	if os.Getenv("SKIP_PACKAGE_UPDATE") != "" {
		glog.Infof("SKIP_PACKAGE_UPDATE was set; skipping package update")
		return nil
	}
	args := []string{"apt-get", "update"}
	glog.Infof("running command %s", args)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error update packages: %v: %s", err, string(output))
	}

	return nil
}

func (_ *UpdatePackages) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *UpdatePackages) error {
	t.Config.PackageUpdate = true
	return nil
}
