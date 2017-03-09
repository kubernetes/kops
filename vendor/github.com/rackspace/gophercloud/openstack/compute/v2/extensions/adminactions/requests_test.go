package adminactions

import (
	"fmt"
	"testing"

	th "github.com/rackspace/gophercloud/testhelper"
	"github.com/rackspace/gophercloud/testhelper/client"
)

const serverID = "adef1234"

func TestCreateBackup(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	mockCreateBackupResponse(t, serverID)

	err := CreateBackup(client.ServiceClient(), serverID, CreateBackupOpts{
		Name:       "Backup 1",
		BackupType: "daily",
		Rotation:   1,
	}).ExtractErr()
	th.AssertNoErr(t, err)
}

func TestCreateBackupNoName(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	mockCreateBackupResponse(t, serverID)

	err := CreateBackup(client.ServiceClient(), serverID, CreateBackupOpts{
		BackupType: "daily",
		Rotation:   1,
	}).ExtractErr()
	if err == nil {
		fmt.Errorf("CreateBackup without a specified Name should throw an Error.")
	}
}

func TestCreateBackupNegativeRotation(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	mockCreateBackupResponse(t, serverID)

	err := CreateBackup(client.ServiceClient(), serverID, CreateBackupOpts{
		Name:       "Backup 1",
		BackupType: "daily",
		Rotation:   -1,
	}).ExtractErr()
	if err == nil {
		fmt.Errorf("CreateBackup without a negative Rotation should throw an Error.")
	}
}

func TestCreateBackupNoType(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	mockCreateBackupResponse(t, serverID)

	err := CreateBackup(client.ServiceClient(), serverID, CreateBackupOpts{
		Name: "Backup 1",

		Rotation: 1,
	}).ExtractErr()
	if err == nil {
		fmt.Errorf("CreateBackup without a specified BackupType should throw an Error.")
	}
}

func TestInjectNetworkInfo(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	mockInjectNetworkInfoResponse(t, serverID)

	err := InjectNetworkInfo(client.ServiceClient(), serverID).ExtractErr()
	th.AssertNoErr(t, err)
}

func TestMigrate(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	mockMigrateResponse(t, serverID)

	err := Migrate(client.ServiceClient(), serverID).ExtractErr()
	th.AssertNoErr(t, err)
}

func TestLiveMigrate(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	mockLiveMigrateResponse(t, serverID)

	err := LiveMigrate(client.ServiceClient(), serverID, LiveMigrateOpts{
		BlockMigration: true,
	}).ExtractErr()
	th.AssertNoErr(t, err)
}

func TestTargetLiveMigrate(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	mockTargetLiveMigrateResponse(t, serverID)

	err := LiveMigrate(client.ServiceClient(), serverID, LiveMigrateOpts{
		Host:           "target-compute",
		BlockMigration: true,
	}).ExtractErr()
	th.AssertNoErr(t, err)
}

func TestResetNetwork(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	mockResetNetworkResponse(t, serverID)

	err := ResetNetwork(client.ServiceClient(), serverID).ExtractErr()
	th.AssertNoErr(t, err)
}

func TestResetState(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	mockResetStateResponse(t, serverID)

	err := ResetState(client.ServiceClient(), serverID, "active").ExtractErr()
	th.AssertNoErr(t, err)
}
