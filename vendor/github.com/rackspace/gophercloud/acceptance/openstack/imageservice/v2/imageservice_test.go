// +build acceptance imageservice

package v2

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/rackspace/gophercloud/acceptance/tools"
	"github.com/rackspace/gophercloud/openstack/imageservice/v2/images"
	"github.com/rackspace/gophercloud/pagination"
	th "github.com/rackspace/gophercloud/testhelper"
)

func TestListImages(t *testing.T) {
	client := newClient(t)

	t.Logf("Id\tName\tOwner\tChecksum\tSizeBytes")

	pager := images.List(client, nil)
	count, pages := 0, 0
	pager.EachPage(func(page pagination.Page) (bool, error) {
		pages++
		t.Logf("---")

		images, err := images.ExtractImages(page)
		if err != nil {
			return false, err
		}

		for _, i := range images {
			t.Logf("%s\t%s\t%s\t%s\t%v\t\n", i.ID, i.Name, i.Owner, i.Checksum, i.SizeBytes)
			count++
		}

		return true, nil
	})

	t.Logf("--------\n%d images listed on %d pages.\n", count, pages)
}

func TestListImagesFilter(t *testing.T) {
	client := newClient(t)
	t.Logf("Id\tName\tOwner\tChecksum\tSizeBytes")

	pager := images.List(client, images.ListOpts{Limit: 1})
	count, pages := 0, 0
	pager.EachPage(func(page pagination.Page) (bool, error) {
		pages++
		t.Logf("---")

		images, err := images.ExtractImages(page)
		if err != nil {
			return false, err
		}

		for _, i := range images {
			t.Logf("%s\t%s\t%s\t%s\t%v\t\n", i.ID, i.Name, i.Owner, i.Checksum, i.SizeBytes)
			count++
		}

		return true, nil
	})

	t.Logf("--------\n%d images listed on %d pages.\n", count, pages)

}

func TestCreateDeleteImage(t *testing.T) {
	client := newClient(t)
	imageName := tools.RandomString("ACCPT", 16)
	containerFormat := "ami"
	createResult := images.Create(client, images.CreateOpts{Name: &imageName,
		ContainerFormat: &containerFormat,
		DiskFormat:      &containerFormat})

	th.AssertNoErr(t, createResult.Err)
	image, err := createResult.Extract()
	th.AssertNoErr(t, err)

	t.Logf("Image %v", image)

	image, err = images.Get(client, image.ID).Extract()
	th.AssertNoErr(t, err)

	th.AssertEquals(t, image.Status, images.ImageStatusQueued)

	deleteResult := images.Delete(client, image.ID)
	th.AssertNoErr(t, deleteResult.Err)
}

func TestUploadDownloadImage(t *testing.T) {
	client := newClient(t)

	//creating image
	imageName := tools.RandomString("ACCPT", 16)
	containerFormat := "ami"
	createResult := images.Create(client, images.CreateOpts{Name: &imageName,
		ContainerFormat: &containerFormat,
		DiskFormat:      &containerFormat})
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
	putImageResult := images.PutImageData(client, image.ID, bytes.NewReader(data))
	th.AssertNoErr(t, putImageResult.Err)

	//checking status
	image, err = images.Get(client, image.ID).Extract()
	th.AssertNoErr(t, err)
	th.AssertEquals(t, image.Status, images.ImageStatusActive)
	th.AssertEquals(t, *image.SizeBytes, 9)

	//downloading image data
	reader, err := images.GetImageData(client, image.ID).Extract()
	th.AssertNoErr(t, err)
	receivedData, err := ioutil.ReadAll(reader)
	t.Logf("Received data %v", receivedData)
	th.AssertNoErr(t, err)
	th.AssertByteArrayEquals(t, data, receivedData)

	//deteting image
	deleteResult := images.Delete(client, image.ID)
	th.AssertNoErr(t, deleteResult.Err)

}

func TestUpdateImage(t *testing.T) {
	client := newClient(t)

	//creating image
	image := createTestImage(t, client)

	t.Logf("Image tags %v", image.Tags)

	tags := []string{"acceptance-testing"}
	updatedImage, err := images.Update(client, image.ID, images.UpdateOpts{
		images.ReplaceImageTags{
			NewTags: tags}}).Extract()
	th.AssertNoErr(t, err)
	t.Logf("Received tags '%v'", tags)
	th.AssertDeepEquals(t, updatedImage.Tags, tags)
}
