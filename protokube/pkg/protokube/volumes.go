package protokube

type Volumes interface {
	AttachVolume(volume *Volume) (string, error)
	FindMountedVolumes() ([]*Volume, error)
	FindMountableVolumes() ([]*Volume, error)
}

type Volume struct {
	Name      string
	Device    string
	Available bool
}
