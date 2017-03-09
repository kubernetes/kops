package volumeactions

import (
	"testing"

	th "github.com/rackspace/gophercloud/testhelper"
	"github.com/rackspace/gophercloud/testhelper/client"
)

func TestAttach(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockAttachResponse(t)

	options := &AttachOpts{
		MountPoint:   "/mnt",
		Mode:         "rw",
		InstanceUUID: "50902f4f-a974-46a0-85e9-7efc5e22dfdd",
	}
	err := Attach(client.ServiceClient(), "cd281d77-8217-4830-be95-9528227c105c", options).ExtractErr()
	th.AssertNoErr(t, err)
}

func TestDetach(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockDetachResponse(t)

	err := Detach(client.ServiceClient(), "cd281d77-8217-4830-be95-9528227c105c").ExtractErr()
	th.AssertNoErr(t, err)
}

func TestReserve(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockReserveResponse(t)

	err := Reserve(client.ServiceClient(), "cd281d77-8217-4830-be95-9528227c105c").ExtractErr()
	th.AssertNoErr(t, err)
}

func TestUnreserve(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockUnreserveResponse(t)

	err := Unreserve(client.ServiceClient(), "cd281d77-8217-4830-be95-9528227c105c").ExtractErr()
	th.AssertNoErr(t, err)
}

func TestInitializeConnection(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockInitializeConnectionResponse(t)

	options := &ConnectorOpts{
		IP:        "127.0.0.1",
		Host:      "stack",
		Initiator: "iqn.1994-05.com.redhat:17cf566367d2",
		Multipath: false,
		Platform:  "x86_64",
		OSType:    "linux2",
	}
	_, err := InitializeConnection(client.ServiceClient(), "cd281d77-8217-4830-be95-9528227c105c", options).Extract()
	th.AssertNoErr(t, err)
}

func TestTerminateConnection(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockTerminateConnectionResponse(t)

	options := &ConnectorOpts{
		IP:        "127.0.0.1",
		Host:      "stack",
		Initiator: "iqn.1994-05.com.redhat:17cf566367d2",
		Multipath: false,
		Platform:  "x86_64",
		OSType:    "linux2",
	}
	err := TerminateConnection(client.ServiceClient(), "cd281d77-8217-4830-be95-9528227c105c", options).ExtractErr()
	th.AssertNoErr(t, err)
}
