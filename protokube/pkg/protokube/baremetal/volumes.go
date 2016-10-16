package baremetal

import (
	"fmt"
	"io/ioutil"
	"k8s.io/kops/protokube/pkg/protokube"
)

type Volumes struct {
	basedir string
}

var _ protokube.Volumes = &Volumes{}

func (v *Volumes) AttachVolume(volume *protokube.Volume) error {

}

func (v *Volumes) FindVolumes() ([]*protokube.Volume, error) {
	files, err := ioutil.ReadDir(v.basedir)
	if err != nil {
		return fmt.Errorf("error reading directory %q: %v", v.basedir, err)
	}

	var volumes []*protokube.Volume
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		volume := &protokube.Volume{}

		volume.ID = file.Name()
		volume.AttachedTo = "me"
	}
}
