package nodetasks

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"k8s.io/kops/upup/pkg/fi/utils"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"time"
)

const (
	systemdSystemPath = "/lib/systemd/system" // TODO: Different on redhat
)

type Service struct {
	Name       string
	Definition *string
	Running    *bool

	// Enabled configures the service to start at boot (or not start at boot)
	Enabled *bool

	ManageState  *bool `json:"manageState"`
	SmartRestart *bool `json:"smartRestart"`
}

var _ fi.HasDependencies = &Service{}

func (p *Service) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	for _, v := range tasks {
		// We assume that services depend on basically everything
		typeName := utils.BuildTypeName(reflect.TypeOf(v))
		switch typeName {
		case "*CopyAssetTask", "*File", "*Package", "*Sysctl", "*UpdatePackages", "*UserTask", "*Disk":
			deps = append(deps, v)
		case "*Service":
		// ignore
		default:
			glog.Warningf("Unhandled type name in Service::GetDependencies: %q", typeName)
			deps = append(deps, v)
		}
	}
	return deps
}

func (s *Service) String() string {
	return fmt.Sprintf("Service: %s", s.Name)
}

func NewService(name string, contents string, meta string) (fi.Task, error) {
	s := &Service{Name: name}
	s.Definition = fi.String(contents)

	if meta != "" {
		err := json.Unmarshal([]byte(meta), s)
		if err != nil {
			return nil, fmt.Errorf("error parsing json for service %q: %v", name, err)
		}
	}

	// Default some values to true: Running, SmartRestart, ManageState
	if s.Running == nil {
		s.Running = fi.Bool(true)
	}
	if s.SmartRestart == nil {
		s.SmartRestart = fi.Bool(true)
	}
	if s.ManageState == nil {
		s.ManageState = fi.Bool(true)
	}

	// Default Enabled to be the same as running
	if s.Enabled == nil {
		s.Enabled = s.Running
	}

	return s, nil
}

func getSystemdStatus(name string) (map[string]string, error) {
	glog.V(2).Infof("querying state of service %q", name)
	cmd := exec.Command("systemctl", "show", "--all", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error doing systemd show %s: %v\nOutput: %s", name, err, output)
	}
	properties := make(map[string]string)
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}
		tokens := strings.SplitN(line, "=", 2)
		if len(tokens) != 2 {
			glog.Warningf("Ignoring line in systemd show output: %q", line)
			continue
		}
		properties[tokens[0]] = tokens[1]
	}
	return properties, nil
}

func (e *Service) Find(c *fi.Context) (*Service, error) {
	servicePath := path.Join(systemdSystemPath, e.Name)

	d, err := ioutil.ReadFile(servicePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("Error reading systemd file %q: %v", servicePath, err)
		}

		// Not found
		return &Service{
			Name:       e.Name,
			Definition: nil,
			Running:    fi.Bool(false),
		}, nil
	}

	actual := &Service{
		Name:       e.Name,
		Definition: fi.String(string(d)),

		// Avoid spurious changes
		ManageState:  e.ManageState,
		SmartRestart: e.SmartRestart,
	}

	properties, err := getSystemdStatus(e.Name)
	if err != nil {
		return nil, err
	}

	activeState := properties["ActiveState"]
	switch activeState {
	case "active":
		actual.Running = fi.Bool(true)

	case "failed", "inactive":
		actual.Running = fi.Bool(false)
	default:
		glog.Warningf("Unknown ActiveState=%q; will treat as not running", activeState)
		actual.Running = fi.Bool(false)
	}

	wantedBy := properties["WantedBy"]
	switch wantedBy {
	case "":
		actual.Enabled = fi.Bool(false)

	// TODO: Can probably do better here!
	case "multi-user.target", "graphical.target multi-user.target":
		actual.Enabled = fi.Bool(true)

	default:
		glog.Warningf("Unknown WantedBy=%q; will treat as not enabled", wantedBy)
		actual.Enabled = fi.Bool(false)
	}

	return actual, nil
}

// Parse the systemd unit file to extract obvious dependencies
func getSystemdDependencies(serviceName string, definition string) ([]string, error) {
	var dependencies []string
	for _, line := range strings.Split(definition, "\n") {
		line = strings.TrimSpace(line)
		tokens := strings.SplitN(line, "=", 2)
		if len(tokens) != 2 {
			continue
		}
		k := strings.TrimSpace(tokens[0])
		v := strings.TrimSpace(tokens[1])
		switch k {
		case "EnvironmentFile":
			dependencies = append(dependencies, v)
		case "ExecStart":
			// ExecStart=/usr/local/bin/kubelet "$DAEMON_ARGS"
			// We extract the first argument (only)
			tokens := strings.SplitN(v, " ", 2)
			dependencies = append(dependencies, tokens[0])
			glog.V(2).Infof("extracted depdendency from %q: %q", line, tokens[0])
		}
	}
	return dependencies, nil
}

func (e *Service) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *Service) CheckChanges(a, e, changes *Service) error {
	return nil
}

func (_ *Service) RenderLocal(t *local.LocalTarget, a, e, changes *Service) error {
	serviceName := e.Name

	action := ""

	if changes.Running != nil && fi.BoolValue(e.ManageState) {
		if fi.BoolValue(e.Running) {
			action = "restart"
		} else {
			action = "stop"
		}
	}

	if changes.Definition != nil {
		servicePath := path.Join(systemdSystemPath, serviceName)
		err := fi.WriteFile(servicePath, fi.NewStringResource(*e.Definition), 0644, 0755)
		if err != nil {
			return fmt.Errorf("error writing systemd service file: %v", err)
		}

		glog.Infof("Reloading systemd configuration")
		cmd := exec.Command("systemctl", "daemon-reload")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error doing systemd daemon-reload: %v\nOutput: %s", err, output)
		}
	}

	// "SmartRestart" - look at the obvious dependencies in the systemd service, restart if start time older
	if fi.BoolValue(e.ManageState) && fi.BoolValue(e.SmartRestart) {
		definition := fi.StringValue(e.Definition)
		if definition == "" && a != nil {
			definition = fi.StringValue(a.Definition)
		}

		if action == "" && fi.BoolValue(e.Running) && definition != "" {
			dependencies, err := getSystemdDependencies(serviceName, definition)
			if err != nil {
				return err
			}

			var newest time.Time
			for _, dependency := range dependencies {
				stat, err := os.Stat(dependency)
				if err != nil {
					glog.Infof("Ignoring error checking service dependency %q: %v", dependency, err)
					continue
				}
				modTime := stat.ModTime()
				if newest.IsZero() || newest.Before(modTime) {
					newest = modTime
				}
			}

			if !newest.IsZero() {
				properties, err := getSystemdStatus(e.Name)
				if err != nil {
					return err
				}

				startedAt := properties["ExecMainStartTimestamp"]
				if startedAt == "" {
					glog.Warningf("service was running, but did not have ExecMainStartTimestamp: %q", serviceName)
				} else {
					startedAtTime, err := time.Parse("Mon 2006-01-02 15:04:05 MST", startedAt)
					if err != nil {
						return fmt.Errorf("unable to parse service ExecMainStartTimestamp: %q", startedAt)
					}
					if startedAtTime.Before(newest) {
						glog.V(2).Infof("will restart service %q because dependency changed after service start", serviceName)
						action = "restart"
					} else {
						glog.V(2).Infof("will not restart service %q - started after dependencies", serviceName)
					}
				}
			}
		}
	}

	if action != "" && fi.BoolValue(e.ManageState) {
		glog.Infof("Restarting service %q", serviceName)
		cmd := exec.Command("systemctl", action, serviceName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error doing systemd %s %s: %v\nOutput: %s", action, serviceName, err, output)
		}
	}

	if changes.Enabled != nil && fi.BoolValue(e.ManageState) {
		var args []string
		if fi.BoolValue(e.Enabled) {
			glog.Infof("Enabling service %q", serviceName)
			args = []string{"enable", serviceName}
		} else {
			glog.Infof("Disabling service %q", serviceName)
			args = []string{"disable", serviceName}
		}
		cmd := exec.Command("systemctl", args...)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error doing 'systemctl %v': %v\nOutput: %s", args, err, output)
		}
	}

	return nil
}

func (_ *Service) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *Service) error {
	serviceName := e.Name

	servicePath := path.Join(systemdSystemPath, serviceName)
	err := t.WriteFile(servicePath, fi.NewStringResource(*e.Definition), 0644, 0755)
	if err != nil {
		return err
	}

	if fi.BoolValue(e.ManageState) {
		t.AddCommand(cloudinit.Once, "systemctl", "daemon-reload")
		t.AddCommand(cloudinit.Once, "systemctl", "start", "--no-block", serviceName)
	}

	return nil
}
