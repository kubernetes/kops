package scalewaymodel

//import (
//	"k8s.io/kops/pkg/apis/kops"
//	"k8s.io/kops/upup/pkg/fi"
//)
//
//type BastionModelBuilder struct {
//	*ScwModelContext
//	Lifecycle fi.Lifecycle
//}
//
//var _ fi.ModelBuilder = &BastionModelBuilder{}
//
//func (b *BastionModelBuilder) Build(c *fi.ModelBuilderContext) error {
//	var bastionInstanceGroups []*kops.InstanceGroup
//	for _, ig := range b.InstanceGroups {
//		if ig.Spec.Role == kops.InstanceGroupRoleBastion {
//			bastionInstanceGroups = append(bastionInstanceGroups, ig)
//		}
//	}
//
//	if len(bastionInstanceGroups) == 0 {
//		return nil
//	}
//
//	// TODO(Mia-Cross): Create security groups
//
//	publicName := ""
//	if b.Cluster.Spec.Topology != nil && b.Cluster.Spec.Topology.Bastion != nil {
//		publicName = b.Cluster.Spec.Topology.Bastion.PublicName
//	}
//	if publicName != "" {
//		// Here we implement the bastion CNAME logic
//		// By default bastions will create a CNAME that follows the `bastion-$clustername` formula
//		t := &scalewaytasks.DNSName{
//			Name:      fi.PtrTo(publicName),
//			Lifecycle: b.Lifecycle,
//
//			Zone:               b.LinkToDNSZone(),
//			ResourceName:       fi.PtrTo(publicName),
//			ResourceType:       fi.PtrTo("A"),
//			TargetLoadBalancer: elb,
//		}
//		c.AddTask(t)
//
//	}
//	return nil
//}
