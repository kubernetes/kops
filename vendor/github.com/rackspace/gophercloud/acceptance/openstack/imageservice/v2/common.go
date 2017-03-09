// +build acceptance imageservice

package v2

import (
	"bytes"
	"os"
	"testing"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/acceptance/tools"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/imageservice/v2/images"
	th "github.com/rackspace/gophercloud/testhelper"
)

func newClient(t *testing.T) *gophercloud.ServiceClient {

	authURL := os.Getenv("OS_AUTH_URL")
	username := os.Getenv("OS_USERNAME")
	password := os.Getenv("OS_PASSWORD")
	tenantName := os.Getenv("OS_TENANT_NAME")
	tenantID := os.Getenv("OS_TENANT_ID")
	domainName := os.Getenv("OS_DOMAIN_NAME")
	regionName := os.Getenv("OS_REGION_NAME")

	t.Logf("Credentials used: OS_AUTH_URL='%s' OS_USERNAME='%s' OS_PASSWORD='*****' OS_TENANT_NAME='%s' OS_TENANT_NAME='%s' OS_REGION_NAME='%s' OS_TENANT_ID='%s' \n",
		authURL, username, tenantName, domainName, regionName, tenantID)

	client, err := openstack.NewClient(authURL)
	th.AssertNoErr(t, err)

	ao := gophercloud.AuthOptions{
		Username:   username,
		Password:   password,
		TenantName: tenantName,
		TenantID:   tenantID,
		DomainName: domainName,
	}

	err = openstack.AuthenticateV3(client, ao)
	th.AssertNoErr(t, err)
	t.Logf("Token is %v", client.TokenID)

	c, err := openstack.NewImageServiceV2(client, gophercloud.EndpointOpts{
		Region: regionName,
	})
	th.AssertNoErr(t, err)
	return c
}

func createTestImage(t *testing.T, client *gophercloud.ServiceClient) images.Image {
	//creating image
	imageName := tools.RandomString("ACCPT", 16)
	containerFormat := "ami"
	createResult := images.Create(client, images.CreateOpts{Name: imageName,
		ContainerFormat: containerFormat,
		DiskFormat:      containerFormat})
	th.AssertNoErr(t, createResult.Err)
	image, err := createResult.Extract()
	th.AssertNoErr(t, err)
	t.Logf("Image %v", image)

	//checking status
	image, err = images.Get(client, image.ID).Extract()
	th.AssertNoErr(t, err)
	th.AssertEquals(t, image.Status, images.ImageStatusQueued)

	//uploading image data
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	uploadResult := images.Upload(client, image.ID, bytes.NewReader(data))
	th.AssertNoErr(t, uploadResult.Err)

	//checking status
	image, err = images.Get(client, image.ID).Extract()
	th.AssertNoErr(t, err)
	th.AssertEquals(t, image.Status, images.ImageStatusActive)
	th.AssertEquals(t, image.SizeBytes, 9)
	return *image
}

func deleteImage(t *testing.T, client *gophercloud.ServiceClient, image images.Image) {
	//deteting image
	deleteResult := images.Delete(client, image.ID)
	th.AssertNoErr(t, deleteResult.Err)
}
