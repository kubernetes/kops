/*
Package volumes provides information and interaction with volumes in the
OpenStack Block Storage service. A volume is a detachable block storage
device, akin to a USB hard drive. It can only be attached to one instance at
a time.

Example of creating Volume B on a Different Host than Volume A

	schedulerHintOpts := volumes.SchedulerHintCreateOpts{
		DifferentHost: []string{
			"volume-a-uuid",
		}
	}

	createOpts := volumes.CreateOpts{
		Name:           "volume_b",
		Size:           10,
	}

	volume, err := volumes.Create(context.TODO(), computeClient, createOpts, schedulerHintOpts).Extract()
	if err != nil {
		panic(err)
	}

Example of creating Volume B on the Same Host as Volume A

	schedulerHintOpts := volumes.SchedulerHintCreateOpts{
		SameHost: []string{
			"volume-a-uuid",
		}
	}

	createOpts := volumes.CreateOpts{
		Name:              "volume_b",
		Size:              10
	}

	volume, err := volumes.Create(context.TODO(), computeClient, createOpts, schedulerHintOpts).Extract()
	if err != nil {
		panic(err)
	}

Example of creating a Volume from a Backup

	backupID := "20c792f0-bb03-434f-b653-06ef238e337e"
	options := volumes.CreateOpts{
		Name:     "vol-001",
		BackupID: &backupID,
	}

	client.Microversion = "3.47"
	volume, err := volumes.Create(context.TODO(), client, options, nil).Extract()
	if err != nil {
		panic(err)
	}

	fmt.Println(volume)

Example of Creating an Image from a Volume

	uploadImageOpts := volumes.UploadImageOpts{
		ImageName: "my_vol",
		Force:     true,
	}

	volumeImage, err := volumes.UploadImage(context.TODO(), client, volume.ID, uploadImageOpts).Extract()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", volumeImage)

Example of Extending a Volume's Size

	extendOpts := volumes.ExtendSizeOpts{
		NewSize: 100,
	}

	err := volumes.ExtendSize(context.TODO(), client, volume.ID, extendOpts).ExtractErr()
	if err != nil {
		panic(err)
	}

Example of Initializing a Volume Connection

	connectOpts := &volumes.InitializeConnectionOpts{
		IP:        "127.0.0.1",
		Host:      "stack",
		Initiator: "iqn.1994-05.com.redhat:17cf566367d2",
		Multipath: gophercloud.Disabled,
		Platform:  "x86_64",
		OSType:    "linux2",
	}

	connectionInfo, err := volumes.InitializeConnection(context.TODO(), client, volume.ID, connectOpts).Extract()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", connectionInfo["data"])

	terminateOpts := &volumes.InitializeConnectionOpts{
		IP:        "127.0.0.1",
		Host:      "stack",
		Initiator: "iqn.1994-05.com.redhat:17cf566367d2",
		Multipath: gophercloud.Disabled,
		Platform:  "x86_64",
		OSType:    "linux2",
	}

	err = volumes.TerminateConnection(context.TODO(), client, volume.ID, terminateOpts).ExtractErr()
	if err != nil {
		panic(err)
	}

Example of Setting a Volume's Bootable status

	options := volumes.BootableOpts{
		Bootable: true,
	}

	err := volumes.SetBootable(context.TODO(), client, volume.ID, options).ExtractErr()
	if err != nil {
		panic(err)
	}

Example of Changing Type of a Volume

	changeTypeOpts := volumes.ChangeTypeOpts{
		NewType:         "ssd",
		MigrationPolicy: volumes.MigrationPolicyOnDemand,
	}

	err = volumes.ChangeType(context.TODO(), client, volumeID, changeTypeOpts).ExtractErr()
	if err != nil {
		panic(err)
	}

Example of Attaching a Volume to an Instance

	attachOpts := volumes.AttachOpts{
		MountPoint:   "/mnt",
		Mode:         "rw",
		InstanceUUID: server.ID,
	}

	err := volumes.Attach(context.TODO(), client, volume.ID, attachOpts).ExtractErr()
	if err != nil {
		panic(err)
	}

	detachOpts := volumes.DetachOpts{
		AttachmentID: volume.Attachments[0].AttachmentID,
	}

	err = volumes.Detach(context.TODO(), client, volume.ID, detachOpts).ExtractErr()
	if err != nil {
		panic(err)
	}
*/
package volumes
