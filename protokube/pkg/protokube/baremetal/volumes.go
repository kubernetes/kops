package baremetal

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/kops/protokube/pkg/protokube"
	"k8s.io/kops/upup/pkg/fi/utils"
	"path"
	"strings"
)

type Volumes struct {
	basedir string
}

var _ protokube.Volumes = &Volumes{}

func NewVolumes(basedir string) (*Volumes, error) {
	v := &Volumes{
		basedir: basedir,
	}
	return v, nil
}

func (v *Volumes) AttachVolume(volume *protokube.Volume) error {
	return nil
}

func (v *Volumes) ClusterID() string {
	return ""
}

func (v *Volumes) FindVolumes() ([]*protokube.Volume, error) {
	files, err := ioutil.ReadDir(v.basedir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory %q: %v", v.basedir, err)
	}

	var volumes []*protokube.Volume
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".meta") {
			continue
		}

		volume := &protokube.Volume{}

		{
			p := path.Join(v.basedir, file.Name())
			data, err := ioutil.ReadFile(p)
			if err != nil {
				glog.Warningf("ignoring error reading file %q: %v", p, err)
				continue
			}

			err = utils.YamlUnmarshal([]byte(data), &volume.Info)
			if err != nil {
				glog.Warningf("ignoring error parsing %q: %v", p, err)
				continue
			}
		}

		id := strings.TrimSuffix(file.Name(), ".meta")

		volume.ID = id
		volume.LocalDevice = path.Join(v.basedir, id)
		volume.AttachedTo = "localhost"
		volume.Mountpoint = path.Join(v.basedir, id)

		volumes = append(volumes, volume)
	}

	return volumes, nil
}
