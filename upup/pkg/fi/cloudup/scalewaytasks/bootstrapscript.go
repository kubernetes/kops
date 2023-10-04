package scalewaytasks

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// +kops:fitask
type BootstrapInstance struct {
	Name *string

	Lifecycle fi.Lifecycle
	Instance  *Instance
	UserData  *fi.Resource
}

var _ fi.CloudupTask = &BootstrapInstance{}
var _ fi.CloudupHasDependencies = &BootstrapInstance{}

func (p *BootstrapInstance) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	for _, task := range tasks {
		if _, ok := task.(*Instance); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*PrivateNIC); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

func (b *BootstrapInstance) Find(c *fi.CloudupContext) (*BootstrapInstance, error) {
	cloud := c.T.Cloud.(scaleway.ScwCloud)

	servers, err := cloud.GetClusterServers(cloud.ClusterName(b.Instance.Tags), b.Name)
	if err != nil {
		return nil, fmt.Errorf("error finding instances: %w", err)
	}
	if len(servers) == 0 {
		return nil, nil
	}

	// Check user-data differences to see if servers updates are needed
	var needsUpdate []string
	for _, server := range servers {
		diff, err := checkUserDataDifferences(c, cloud, server, b.UserData)
		if scaleway.Is404Error(err) {
			// No user-data set up on the server yet
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("checking user-data differences in server %s (%s): %w", server.Name, server.ID, err)
		}
		if diff == true {
			needsUpdate = append(needsUpdate, server.ID)
		}
	}
	b.Instance.NeedsUpdate = append(b.Instance.NeedsUpdate, needsUpdate...)

	return &BootstrapInstance{
		Name:      b.Name,
		Instance:  b.Instance,
		Lifecycle: b.Lifecycle,
		// server.user-data plutot non ??
		UserData: b.UserData,
	}, nil

}

func (b *BootstrapInstance) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(b, c)
}

func (_ *BootstrapInstance) CheckChanges(actual, expected, changes *BootstrapInstance) error {
	return nil
}

func (_ *BootstrapInstance) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *BootstrapInstance) error {
	cloud := t.Cloud.(scaleway.ScwCloud)
	instanceService := cloud.InstanceService()
	clusterName := scaleway.ClusterNameFromTags(expected.Instance.Tags)
	zone := scw.Zone(fi.ValueOf(expected.Instance.Zone))
	igName := fi.ValueOf(expected.Name)

	// We load the cloud-init script in the instance user data
	userData, err := fi.ResourceAsBytes(*expected.UserData)
	if err != nil {
		return fmt.Errorf("rendering bootstrapscript for instance group %q: %w", igName, err)
	}

	servers, err := cloud.GetClusterServers(clusterName, &igName)
	if err != nil {
		return fmt.Errorf("rendering bootstrapscript for instance group %q: getting servers: %w", igName, err)
	}

	for _, server := range servers {

		err = instanceService.SetServerUserData(&instance.SetServerUserDataRequest{
			ServerID: server.ID,
			Zone:     server.Zone,
			Key:      "cloud-init",
			Content:  bytes.NewBuffer(userData),
		})
		if err != nil {
			return fmt.Errorf("error setting 'cloud-init' in user-data for instance %s of group %q: %w", server.ID, igName, err)
		}

		// We start the instance
		_, err = instanceService.ServerAction(&instance.ServerActionRequest{
			Zone:     zone,
			ServerID: server.ID,
			Action:   instance.ServerActionPoweron,
		})
		if err != nil {
			return fmt.Errorf("error powering on instance %s of group %q: %w", server.ID, igName, err)
		}

		// We wait for the instance to be ready
		_, err = instanceService.WaitForServer(&instance.WaitForServerRequest{
			ServerID: server.ID,
			Zone:     zone,
		})
		if err != nil {
			return fmt.Errorf("error waiting for instance %s of group %q: %w", server.ID, igName, err)
		}
	}
	return nil
}

func (_ *BootstrapInstance) RenderTerraform(t *terraform.TerraformTarget, actual, expected, changes *BootstrapInstance) error {
	//TODO(Mia-Cross): implement me
	panic("RenderTerraform is not implemented")
}

func checkUserDataDifferences(c *fi.CloudupContext, cloud scaleway.ScwCloud, actualServer *instance.Server, expectedUserData *fi.Resource) (bool, error) {
	actualUserData, err := cloud.InstanceService().GetServerUserData(&instance.GetServerUserDataRequest{
		Zone:     actualServer.Zone,
		ServerID: actualServer.ID,
		Key:      "cloud-init",
	}, scw.WithContext(c.Context()))
	if err != nil {
		return false, fmt.Errorf("getting actual user-data: %w", err)
	}

	actualUserDataBytes, err := io.ReadAll(actualUserData)
	if err != nil {
		return false, fmt.Errorf("reading actual user-data: %w", err)
	}
	expectedUserDataBytes, err := fi.ResourceAsBytes(*expectedUserData)
	if err != nil {
		return false, fmt.Errorf("reading expected user-data: %w", err)
	}

	if sha256.Sum256(actualUserDataBytes) != sha256.Sum256(expectedUserDataBytes) {
		return true, nil
	}
	return false, nil
}
