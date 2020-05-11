/*
Package volumes provides information and interaction with volumes in the
OpenStack Block Storage service. A volume is a detachable block storage
device, akin to a USB hard drive. It can only be attached to one instance at
a time.

Example to create a Volume from a Backup

	backupID := "20c792f0-bb03-434f-b653-06ef238e337e"
	options := volumes.CreateOpts{
		Name:     "vol-001",
		BackupID: &backupID,
	}

	client.Microversion = "3.47"
	volume, err := volumes.Create(client, options).Extract()
	if err != nil {
		panic(err)
	}

	fmt.Println(volume)
*/
package volumes
