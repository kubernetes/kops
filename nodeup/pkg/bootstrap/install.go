package bootstrap

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kubernetes/pkg/util/sets"
	"strings"
	"time"
)

type Installation struct {
	FSRoot          string
	CacheDir        string
	MaxTaskDuration time.Duration
	Command         []string
}

func (i *Installation) Run() error {
	distribution, err := distros.FindDistribution(i.FSRoot)
	if err != nil {
		return fmt.Errorf("error determining OS distribution: %v", err)
	}

	tags := sets.NewString()
	tags.Insert(distribution.BuildTags()...)

	tasks := make(map[string]fi.Task)

	buildContext := &fi.ModelBuilderContext{
		Tasks: tasks,
	}
	i.Build(buildContext)

	// If there is a package task, we need an update packages task
	for _, t := range tasks {
		if _, ok := t.(*nodetasks.Package); ok {
			glog.Infof("Package task found; adding UpdatePackages task")
			tasks["UpdatePackages"] = nodetasks.NewUpdatePackages()
			break
		}
	}

	if tasks["UpdatePackages"] == nil {
		glog.Infof("No package task found; won't update packages")
	}

	var configBase vfs.Path
	var cloud fi.Cloud
	var keyStore fi.Keystore
	var secretStore fi.SecretStore

	target := &local.LocalTarget{
		CacheDir: i.CacheDir,
		Tags:     tags,
	}

	checkExisting := true
	context, err := fi.NewContext(target, cloud, keyStore, secretStore, configBase, checkExisting, tasks)
	if err != nil {
		return fmt.Errorf("error building context: %v", err)
	}
	defer context.Close()

	err = context.RunTasks(i.MaxTaskDuration)
	if err != nil {
		return fmt.Errorf("error running tasks: %v", err)
	}

	err = target.Finish(tasks)
	if err != nil {
		return fmt.Errorf("error finishing target: %v", err)
	}

	return nil
}
func (i *Installation) Build(c *fi.ModelBuilderContext) {
	c.AddTask(i.buildSystemdJob())
}

func (i *Installation) buildSystemdJob() *nodetasks.Service {
	command := strings.Join(i.Command, " ")

	serviceName := "kops-configuration.service"

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Run kops bootstrap (nodeup)")
	manifest.Set("Unit", "Documentation", "https://github.com/kubernetes/kops")

	manifest.Set("Service", "ExecStart", command)
	manifest.Set("Service", "Type", "oneshot")

	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	glog.V(8).Infof("Built service manifest %q\n%s", serviceName, manifestString)

	service := &nodetasks.Service{
		Name:       serviceName,
		Definition: fi.String(manifestString),
	}

	service.InitDefaults()

	return service
}
