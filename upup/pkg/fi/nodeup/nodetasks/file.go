package nodetasks

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kube-deploy/upup/pkg/fi/nodeup/local"
	"os"
	"os/exec"
	"strings"
)

type File struct {
	Path     string
	Contents fi.Resource

	Mode        *string `json:"mode"`
	IfNotExists bool    `json:"ifNotExists"`

	OnChangeExecute []string `json:"onChangeExecute,omitempty"`
}

var _ fi.Task = &File{}

func NewFileTask(name string, src fi.Resource, destPath string, meta string) (*File, error) {
	f := &File{
		//Name:     name,
		Contents: src,
		Path:     destPath,
	}

	if meta != "" {
		err := json.Unmarshal([]byte(meta), f)
		if err != nil {
			return nil, fmt.Errorf("error parsing meta for file %q: %v", name, err)
		}
	}

	return f, nil
}

func (f *File) String() string {
	return fmt.Sprintf("File: %q", f.Path)
}

func findFile(p string) (*File, error) {
	stat, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
	}

	actual := &File{}
	actual.Path = p
	actual.Mode = fi.String(fi.FileModeToString(stat.Mode()))
	actual.Contents = fi.NewFileResource(p)

	return actual, nil
}

func (e *File) Find(c *fi.Context) (*File, error) {
	actual, err := findFile(e.Path)
	if err != nil {
		return nil, err
	}
	if actual == nil {
		return nil, nil
	}

	// To avoid spurious changes
	actual.IfNotExists = e.IfNotExists

	return actual, nil
}

func (e *File) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *File) CheckChanges(a, e, changes *File) error {
	return nil
}

func (_ *File) RenderLocal(t *local.LocalTarget, a, e, changes *File) error {
	dirMode := os.FileMode(0755)
	fileMode, err := fi.ParseFileMode(fi.StringValue(e.Mode), 0644)
	if err != nil {
		return fmt.Errorf("invalid file mode for %q: %q", e.Path, e.Mode)
	}

	if a != nil {
		if e.IfNotExists {
			glog.V(2).Infof("file exists and IfNotExists set; skipping %q", e.Path)
			return nil
		}
	}

	changed := false
	if changes.Contents != nil {
		err = fi.WriteFile(e.Path, e.Contents, fileMode, dirMode)
		if err != nil {
			return fmt.Errorf("error copying file %q: %v", e.Path, err)
		}
		changed = true
	} else if changes.Mode != nil {
		modeChanged, err := fi.EnsureFileMode(e.Path, fileMode)
		if err != nil {
			return fmt.Errorf("error changing file mode %q: %v", e.Path, err)
		}
		changed = changed || modeChanged
	}

	if changed && e.OnChangeExecute != nil {
		args := e.OnChangeExecute
		human := strings.Join(args, " ")

		glog.Infof("Changed; will execute OnChangeExecute command: %q", human)

		cmd := exec.Command(args[0], args[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error executing command %q: %v\nOutput: %s", human, err, output)
		}
	}

	return nil
}

func (_ *File) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *File) error {
	dirMode := os.FileMode(0755)
	fileMode, err := fi.ParseFileMode(fi.StringValue(e.Mode), 0644)
	if err != nil {
		return fmt.Errorf("invalid file mode for %q: %q", e.Path, e.Mode)
	}

	err = t.WriteFile(e.Path, e.Contents, fileMode, dirMode)
	if err != nil {
		return err
	}

	if e.OnChangeExecute != nil {
		t.AddCommand(cloudinit.Always, e.OnChangeExecute...)
	}

	return nil
}
