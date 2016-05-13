package nodetasks

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kube-deploy/upup/pkg/fi/nodeup/local"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
	"os/exec"
)

// UserTask is responsible for creating a user, by calling useradd
type UserTask struct {
	Name string

	Shell string `json:"shell"`
	Home  string `json:"home"`
}

var _ fi.Task = &UserTask{}

func (e *UserTask) String() string {
	return fmt.Sprintf("User: %s", e.Name)
}

func NewUserTask(name string, contents string, meta string) (fi.Task, error) {
	s := &UserTask{Name: name}

	err := utils.YamlUnmarshal([]byte(contents), s)
	if err != nil {
		return nil, fmt.Errorf("error parsing json for service %q: %v", name, err)
	}

	return s, nil
}

func (e *UserTask) Find(c *fi.Context) (*UserTask, error) {
	info, err := fi.LookupUser(e.Name)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, nil
	}

	actual := &UserTask{
		Name:  e.Name,
		Shell: info.Shell,
		Home:  info.Home,
	}

	return actual, nil
}

func (e *UserTask) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *UserTask) CheckChanges(a, e, changes *UserTask) error {
	return nil
}

func buildUseraddArgs(e *UserTask) []string {
	var args []string
	if e.Shell != "" {
		args = append(args, "-s", e.Shell)
	}
	if e.Home != "" {
		args = append(args, "-d", e.Home)
	}
	args = append(args, e.Name)
	return args
}

func (_ *UserTask) RenderLocal(t *local.LocalTarget, a, e, changes *UserTask) error {
	if a == nil {
		args := buildUseraddArgs(e)
		glog.Infof("Creating user %q", e.Name)
		cmd := exec.Command("useradd", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error creating user: %v\nOutput: %s", err, output)
		}
	} else {
		var args []string

		if changes.Shell != "" {
			args = append(args, "-s", e.Shell)
		}
		if changes.Home != "" {
			args = append(args, "-d", e.Home)
		}

		if len(args) != 0 {
			args = append(args, e.Name)
			glog.Infof("Reconfiguring user %q", e.Name)
			cmd := exec.Command("usermod", args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("error reconfiguring user: %v\nOutput: %s", err, output)
			}
		}
	}

	return nil
}

func (_ *UserTask) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *UserTask) error {
	args := buildUseraddArgs(e)
	cmd := []string{"useradd"}
	cmd = append(cmd, args...)
	glog.Infof("Creating user %q", e.Name)
	t.AddCommand(cloudinit.Once, cmd...)

	return nil
}
