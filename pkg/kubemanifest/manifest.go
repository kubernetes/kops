package kubemanifest

import (
	"bytes"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/golang/glog"
)

type Manifest struct {
	//bytes []byte
	data map[string]interface{}
}

func LoadManifestsFrom(contents []byte) ([]*Manifest, error) {
	var manifests []*Manifest

	// TODO: Support more separators?
	sections := bytes.Split(contents, []byte("\n---\n"))

	for _, section := range sections {
		data := make(map[string]interface{})
		err := yaml.Unmarshal(section, &data)
		if err != nil {
			return nil, fmt.Errorf("error parsing yaml: %v", err)
		}

		manifest := &Manifest{
			//bytes: section,
			data: data,
		}
		manifests = append(manifests, manifest)
	}

	return manifests, nil
}

func (m *Manifest) ToYAML() ([]byte, error) {
	//if m.bytes == nil {
	b, err := yaml.Marshal(m.data)
	if err != nil {
		return nil, fmt.Errorf("error marshalling manifest to yaml: %v", err)
	}
	//	m.bytes = b
	//}
	//return m.bytes, nil
	return b, nil
}

func (m *Manifest) accept(visitor Visitor) error {
	err := visit(visitor, m.data, []string{}, func(v interface{}) {
		glog.Fatalf("cannot mutate top-level data")
	})
	return err
}
