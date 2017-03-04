package vspheremodel

import (
	"k8s.io/kops/upup/pkg/fi"
	"fmt"
)

// Do we need this model builder?

// AutoscalingGroupModelBuilder configures AutoscalingGroup objects
type VirtualMachineModelBuilder struct {
	*VSphereModelContext
}

func (b *VirtualMachineModelBuilder) Build(c *fi.ModelBuilderContext) error {
	fmt.Print("In VirtualMachineModelBuilder.Build function!!")
	return nil
}
