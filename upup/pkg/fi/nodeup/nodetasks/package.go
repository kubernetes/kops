package nodetasks

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/hashing"
	"k8s.io/kube-deploy/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kube-deploy/upup/pkg/fi/nodeup/local"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Package struct {
	Name string

	Version      *string `json:"version"`
	Source       *string `json:"source"`
	Hash         *string `json:"hash"`
	PreventStart *bool   `json:"preventStart"`
}

const (
	localPackageDir = "/var/cache/nodeup/packages/"
)

var _ fi.HasDependencies = &Package{}

func (p *Package) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	for _, v := range tasks {
		if _, ok := v.(*UpdatePackages); ok {
			deps = append(deps, v)
		}
	}
	return deps
}

func (p *Package) String() string {
	return fmt.Sprintf("Package: %s", p.Name)
}

func NewPackage(name string, contents string, meta string) (fi.Task, error) {
	p := &Package{Name: name}
	if contents != "" {
		err := json.Unmarshal([]byte(contents), p)
		if err != nil {
			return nil, fmt.Errorf("error parsing json for package %q: %v", name, err)
		}
	}
	return p, nil
}

func (e *Package) Find(c *fi.Context) (*Package, error) {
	args := []string{"dpkg-query", "-f", "${db:Status-Abbrev}${Version}\\n", "-W", e.Name}
	human := strings.Join(args, " ")

	glog.V(2).Infof("Listing installed packages: %s", human)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "no packages found") {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing installed packages: %v: %s", err, string(output))
	}

	installed := false
	installedVersion := ""
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}

		tokens := strings.Split(line, " ")
		if len(tokens) != 2 {
			return nil, fmt.Errorf("error parsing dpkg-query line %q", line)
		}
		state := tokens[0]
		version := tokens[1]

		switch state {
		case "ii":
			installed = true
			installedVersion = version
		case "rc":
			// removed
			installed = false
		case "un":
			// unknown
			installed = false
		default:
			return nil, fmt.Errorf("unknown package state %q in line %q", state, line)
		}
	}

	if !installed {
		return nil, nil
	}

	return &Package{
		Name:    e.Name,
		Version: fi.String(installedVersion),
	}, nil
}

func (e *Package) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *Package) CheckChanges(a, e, changes *Package) error {
	return nil
}

func (_ *Package) RenderLocal(t *local.LocalTarget, a, e, changes *Package) error {
	if changes.Version != nil {
		glog.Infof("Installing package %q", e.Name)

		if e.Source != nil {
			// Install a deb
			local := path.Join(localPackageDir, e.Name)
			err := os.MkdirAll(localPackageDir, 0755)
			if err != nil {
				return fmt.Errorf("error creating directories %q: %v", path.Dir(local), err)
			}

			var hash *hashing.Hash
			if fi.StringValue(e.Hash) != "" {
				parsed, err := hashing.FromString(fi.StringValue(e.Hash))
				if err != nil {
					return fmt.Errorf("error paring hash: %v", err)
				}
				hash = parsed
			}
			_, err = fi.DownloadURL(fi.StringValue(e.Source), local, hash)
			if err != nil {
				return err
			}

			args := []string{"dpkg", "-i", local}
			glog.Infof("running command %s", args)
			cmd := exec.Command(args[0], args[1:]...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("error installing package %q: %v: %s", e.Name, err, string(output))
			}
		} else {
			args := []string{"apt-get", "install", "--yes", e.Name}
			glog.Infof("running command %s", args)
			cmd := exec.Command(args[0], args[1:]...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("error installing package %q: %v: %s", e.Name, err, string(output))
			}
		}
	}

	return nil
}

func (_ *Package) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *Package) error {
	if e.Source != nil {

		localFile := path.Join(localPackageDir, e.Name)
		t.AddMkdirpCommand(localPackageDir, 0755)

		url := *e.Source
		t.AddDownloadCommand(cloudinit.Always, url, localFile)

		t.AddCommand(cloudinit.Always, "dpkg", "-i", localFile)
	} else {
		packageSpec := e.Name
		if e.Version != nil {
			packageSpec += " " + *e.Version
		}
		t.Config.Packages = append(t.Config.Packages, packageSpec)
	}

	return nil
}
