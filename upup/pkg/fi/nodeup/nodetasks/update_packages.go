package nodetasks

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"k8s.io/kops/upup/pkg/fi/nodeup/tags"
	"os"
	"os/exec"
)

type UpdatePackages struct {
	// We can't be completely empty or we don't run
	Updated bool
}

var _ fi.HasDependencies = &UpdatePackages{}

func NewUpdatePackages() *UpdatePackages {
	return &UpdatePackages{Updated: true}
}

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

func (s *UpdatePackages) CheckChanges(a, e, changes *UpdatePackages) error {
	return nil
}

func (_ *UpdatePackages) RenderLocal(t *local.LocalTarget, a, e, changes *UpdatePackages) error {
	if os.Getenv("SKIP_PACKAGE_UPDATE") != "" {
		glog.Infof("SKIP_PACKAGE_UPDATE was set; skipping package update")
		return nil
	}
	var args []string
	if t.HasTag(tags.TagOSFamilyDebian) {
		args = []string{"apt-get", "update"}

	} else if t.HasTag(tags.TagOSFamilyRHEL) {
		// Probably not technically needed
		args = []string{"/usr/bin/yum", "check-update"}
	} else {
		return fmt.Errorf("unsupported package system")
	}
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
