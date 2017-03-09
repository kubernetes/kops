package volumes

import (
	"testing"

	"github.com/rackspace/gophercloud/pagination"
	th "github.com/rackspace/gophercloud/testhelper"
	"github.com/rackspace/gophercloud/testhelper/client"
)

func TestList(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockListResponse(t)

	count := 0

	List(client.ServiceClient(), &ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		count++
		actual, err := ExtractVolumes(page)
		if err != nil {
			t.Errorf("Failed to extract volumes: %v", err)
			return false, err
		}

		expected := []Volume{
			{
				ID:   "289da7f8-6440-407c-9fb4-7db01ec49164",
				Name: "vol-001",
				Attachments: []map[string]interface{}{{
					"AttachmentID": "03987cd1-0ad5-40d1-9b2a-7cc48295d4fa",
					"ID":           "47e9ecc5-4045-4ee3-9a4b-d859d546a0cf",
					"VolumeID":     "289da7f8-6440-407c-9fb4-7db01ec49164",
					"ServerID":     "d1c4788b-9435-42e2-9b81-29f3be1cd01f",
					"HostName":     "stack",
					"Device":       "/dev/vdc",
				}},
				AvailabilityZone:          "nova",
				Bootable:                  "false",
				ConsistencyGroupID:        "",
				CreatedAt:                 "2015-09-17T03:35:03.000000",
				Description:               "",
				Encrypted:                 false,
				Metadata:                  map[string]string{"foo": "bar"},
				Multiattach:               false,
				TenantID:                  "304dc00909ac4d0da6c62d816bcb3459",
				ReplicationDriverData:     "",
				ReplicationExtendedStatus: "",
				ReplicationStatus:         "disabled",
				Size:                      75,
				SnapshotID:                "",
				SourceVolID:               "",
				Status:                    "available",
				UserID:                    "ff1ce52c03ab433aaba9108c2e3ef541",
				VolumeType:                "lvmdriver-1",
			},
			{
				ID:                        "96c3bda7-c82a-4f50-be73-ca7621794835",
				Name:                      "vol-002",
				Attachments:               []map[string]interface{}{},
				AvailabilityZone:          "nova",
				Bootable:                  "false",
				ConsistencyGroupID:        "",
				CreatedAt:                 "2015-09-17T03:32:29.000000",
				Description:               "",
				Encrypted:                 false,
				Metadata:                  map[string]string{},
				Multiattach:               false,
				TenantID:                  "304dc00909ac4d0da6c62d816bcb3459",
				ReplicationDriverData:     "",
				ReplicationExtendedStatus: "",
				ReplicationStatus:         "disabled",
				Size:                      75,
				SnapshotID:                "",
				SourceVolID:               "",
				Status:                    "available",
				UserID:                    "ff1ce52c03ab433aaba9108c2e3ef541",
				VolumeType:                "lvmdriver-1",
			},
		}

		th.CheckDeepEquals(t, expected, actual)

		return true, nil
	})

	if count != 1 {
		t.Errorf("Expected 1 page, got %d", count)
	}
}

func TestListAll(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockListResponse(t)

	allPages, err := List(client.ServiceClient(), &ListOpts{}).AllPages()
	th.AssertNoErr(t, err)
	actual, err := ExtractVolumes(allPages)
	th.AssertNoErr(t, err)

	expected := []Volume{
		{
			ID:   "289da7f8-6440-407c-9fb4-7db01ec49164",
			Name: "vol-001",
			Attachments: []map[string]interface{}{{
				"AttachmentID": "03987cd1-0ad5-40d1-9b2a-7cc48295d4fa",
				"ID":           "47e9ecc5-4045-4ee3-9a4b-d859d546a0cf",
				"VolumeID":     "289da7f8-6440-407c-9fb4-7db01ec49164",
				"ServerID":     "d1c4788b-9435-42e2-9b81-29f3be1cd01f",
				"HostName":     "stack",
				"Device":       "/dev/vdc",
			}},
			AvailabilityZone:          "nova",
			Bootable:                  "false",
			ConsistencyGroupID:        "",
			CreatedAt:                 "2015-09-17T03:35:03.000000",
			Description:               "",
			Encrypted:                 false,
			Metadata:                  map[string]string{"foo": "bar"},
			Multiattach:               false,
			TenantID:                  "304dc00909ac4d0da6c62d816bcb3459",
			ReplicationDriverData:     "",
			ReplicationExtendedStatus: "",
			ReplicationStatus:         "disabled",
			Size:                      75,
			SnapshotID:                "",
			SourceVolID:               "",
			Status:                    "available",
			UserID:                    "ff1ce52c03ab433aaba9108c2e3ef541",
			VolumeType:                "lvmdriver-1",
		},
		{
			ID:                        "96c3bda7-c82a-4f50-be73-ca7621794835",
			Name:                      "vol-002",
			Attachments:               []map[string]interface{}{},
			AvailabilityZone:          "nova",
			Bootable:                  "false",
			ConsistencyGroupID:        "",
			CreatedAt:                 "2015-09-17T03:32:29.000000",
			Description:               "",
			Encrypted:                 false,
			Metadata:                  map[string]string{},
			Multiattach:               false,
			TenantID:                  "304dc00909ac4d0da6c62d816bcb3459",
			ReplicationDriverData:     "",
			ReplicationExtendedStatus: "",
			ReplicationStatus:         "disabled",
			Size:                      75,
			SnapshotID:                "",
			SourceVolID:               "",
			Status:                    "available",
			UserID:                    "ff1ce52c03ab433aaba9108c2e3ef541",
			VolumeType:                "lvmdriver-1",
		},
	}

	th.CheckDeepEquals(t, expected, actual)

}

func TestGet(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockGetResponse(t)

	v, err := Get(client.ServiceClient(), "d32019d3-bc6e-4319-9c1d-6722fc136a22").Extract()
	th.AssertNoErr(t, err)

	th.AssertEquals(t, v.Name, "vol-001")
	th.AssertEquals(t, v.ID, "d32019d3-bc6e-4319-9c1d-6722fc136a22")
}

func TestCreate(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockCreateResponse(t)

	options := &CreateOpts{Size: 75, Name: "vol-001"}
	n, err := Create(client.ServiceClient(), options).Extract()
	th.AssertNoErr(t, err)

	th.AssertEquals(t, n.Size, 75)
	th.AssertEquals(t, n.ID, "d32019d3-bc6e-4319-9c1d-6722fc136a22")
}

func TestDelete(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockDeleteResponse(t)

	res := Delete(client.ServiceClient(), "d32019d3-bc6e-4319-9c1d-6722fc136a22")
	th.AssertNoErr(t, res.Err)
}

func TestUpdate(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockUpdateResponse(t)

	options := UpdateOpts{Name: "vol-002"}
	v, err := Update(client.ServiceClient(), "d32019d3-bc6e-4319-9c1d-6722fc136a22", options).Extract()
	th.AssertNoErr(t, err)
	th.CheckEquals(t, "vol-002", v.Name)
}
