package nodetasks

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/hashing"
	"k8s.io/kops/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"sync"
)

type Package struct {
	Name string

	Version      *string `json:"version"`
	Source       *string `json:"source"`
	Hash         *string `json:"hash"`
	PreventStart *bool   `json:"preventStart"`

	// Healthy is true if the package installation did not fail
	Healthy *bool `json:"healthy"`
}

const (
	localPackageDir = "/var/cache/nodeup/packages/"
)

var _ fi.HasDependencies = &Package{}

// GetDependencies computes dependencies for the package task
func (p *Package) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task

	// UpdatePackages before we install any packages
	for _, v := range tasks {
		if _, ok := v.(*UpdatePackages); ok {
			deps = append(deps, v)
		}
	}

	// If this package is a bare deb, install it after OS managed packages
	if !p.isOSPackage() {
		for _, v := range tasks {
			if vp, ok := v.(*Package); ok {
				if vp.isOSPackage() {
					deps = append(deps, v)
				}
			}
		}
	}

	return deps
}

// isOSPackage returns true if this is an OS provided package (as opposed to a bare .deb, for example)
func (p *Package) isOSPackage() bool {
	return fi.StringValue(p.Source) == ""
}

// String returns a string representation, implementing the Stringer interface
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

	// Default values: we want to install a package so that it is healthy
	if p.Healthy == nil {
		p.Healthy = fi.Bool(true)
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
	var healthy *bool
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
			healthy = fi.Bool(true)
		case "iF":
			installed = true
			installedVersion = version
			healthy = fi.Bool(false)
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
		Healthy: healthy,
	}, nil
}

func (e *Package) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *Package) CheckChanges(a, e, changes *Package) error {
	return nil
}

// packageManagerLock is a simple lock that prevents concurrent package manager operations
// It just avoids unnecessary failures from running e.g. concurrent apt-get installs
var packageManagerLock sync.Mutex

func (_ *Package) RenderLocal(t *local.LocalTarget, a, e, changes *Package) error {
	packageManagerLock.Lock()
	defer packageManagerLock.Unlock()

	if a == nil || changes.Version != nil {
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
	} else {
		if changes.Healthy != nil {
			args := []string{"dpkg", "--configure", "-a"}
			glog.Infof("package is not healthy; runnning command %s", args)
			cmd := exec.Command(args[0], args[1:]...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("error running `dpkg --configure -a`: %v: %s", err, string(output))
			}

			changes.Healthy = nil
		}

		if !reflect.DeepEqual(changes, &Package{}) {
			glog.Warningf("cannot apply package changes for %q: %v", e.Name, changes)
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
