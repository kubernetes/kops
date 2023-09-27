package scalewaytasks

//import (
//	"fmt"
//
//	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
//	"github.com/scaleway/scaleway-sdk-go/scw"
//	"k8s.io/kops/upup/pkg/fi"
//	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
//)
//
//type PrivateNIC struct {
//	ID   *string
//	Name *string
//	Zone *string
//	Tags []string
//
//	InstanceID *string
//
//	Lifecycle      fi.Lifecycle
//	PrivateNetwork *PrivateNetwork
//}
//
//var _ fi.CloudupTask = &PrivateNIC{}
//var _ fi.CompareWithID = &PrivateNIC{}
//var _ fi.CloudupHasDependencies = &PrivateNIC{}
//
//func (p *PrivateNIC) CompareWithID() *string {
//	return p.Name
//}
//
//func (p *PrivateNIC) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
//	var deps []fi.CloudupTask
//	for _, task := range tasks {
//		if _, ok := task.(*Instance); ok {
//			deps = append(deps, task)
//		}
//		if _, ok := task.(*PrivateNetwork); ok {
//			deps = append(deps, task)
//		}
//	}
//	return deps
//}
//
//func (p *PrivateNIC) Find(context *fi.CloudupContext) (*PrivateNIC, error) {
//cloud := context.T.Cloud.(scaleway.ScwCloud)
//dhcps, err := cloud.InstanceService().ListPrivateNICs(&instance.ListPrivateNICsRequest{
//	Zone:     scw.Zone(cloud.Zone()),
//	ServerID: p.Instance.,
//	Tags:     nil,
//	PerPage:  nil,
//	Page:     nil,
//})
//if err != nil {
//
//}
//	return &PrivateNIC{}, err
//}
//
//func (p *PrivateNIC) Run(context *fi.CloudupContext) error {
//	return fi.CloudupDefaultDeltaRunMethod(p, context)
//}
//
//func (p *PrivateNIC) CheckChanges(actual, expected, changes *PrivateNIC) error {
//	if actual != nil {
//		if changes.Name != nil {
//			return fi.CannotChangeField("Name")
//		}
//		if changes.Zone != nil {
//			return fi.CannotChangeField("Zone")
//		}
//	} else {
//		if expected.Name == nil {
//			return fi.RequiredField("Name")
//		}
//		if expected.Zone == nil {
//			return fi.RequiredField("Zone")
//		}
//		if expected.InstanceID == nil {
//			return fi.RequiredField("InstanceID")
//		}
//	}
//}
//
//func (_ *PrivateNIC) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *PrivateNIC) error {
//	cloud := t.Cloud.(scaleway.ScwCloud)
//	zone := scw.Zone(fi.ValueOf(expected.Zone))
//
//	if actual != nil {
//		return nil
//	}
//	pNICCreated, err := cloud.InstanceService().CreatePrivateNIC(&instance.CreatePrivateNICRequest{
//		Zone:             zone,
//		ServerID:         fi.ValueOf(expected.InstanceID),
//		PrivateNetworkID: fi.ValueOf(expected.PrivateNetwork.ID),
//	})
//	if err != nil {
//		return fmt.Errorf("creating private NIC between instance %s and private network %s: %w", fi.ValueOf(expected.InstanceID), fi.ValueOf(expected.PrivateNetwork.ID), err)
//	}
//
//	// We wait for the private nic to be ready before proceeding
//	_, err = cloud.InstanceService().WaitForPrivateNIC(&instance.WaitForPrivateNICRequest{
//		ServerID:     fi.ValueOf(expected.InstanceID),
//		PrivateNicID: pNICCreated.PrivateNic.ID,
//		Zone:         zone,
//	})
//	if err != nil {
//		return fmt.Errorf("waiting for private NIC %s: %w", pNICCreated.PrivateNic.ID, err)
//	}
//	return nil
//}
