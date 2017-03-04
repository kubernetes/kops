package vspheremodel

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/pkg/model"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi/cloudup/vspheretasks"
)

// AutoscalingGroupModelBuilder configures AutoscalingGroup objects
type AutoscalingGroupModelBuilder struct {
	*VSphereModelContext

	BootstrapScript *model.BootstrapScript
}

var _ fi.ModelBuilder = &AutoscalingGroupModelBuilder{}

func (b *AutoscalingGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	glog.Warning("AutoscalingGroupModelBuilder.Build not implemented for vsphere")
	for _, ig := range b.InstanceGroups {
		name := b.AutoscalingGroupName(ig)
		createVmTask := &vspheretasks.VirtualMachine{
			Name: &name,
			VMTemplateName: fi.String("dummyVmTemplate"),
		}
		c.AddTask(createVmTask)

		attachISOTaskName := "AttachISO-" + name
		attachISOTask := &vspheretasks.AttachISO{
			Name: &attachISOTaskName,
			VM: createVmTask,
		}
		c.AddTask(attachISOTask)

		powerOnTaskName := "PowerON-" + name
		powerOnTask := &vspheretasks.VMPowerOn{
			Name: &powerOnTaskName,
			AttachISO: attachISOTask,
		}
		c.AddTask(powerOnTask)
	}
	return nil
}
