package utils

import (
	//"gopkg.in/yaml.v2"
	"github.com/ghodss/yaml"
)

// See http://ghodss.com/2014/the-right-way-to-handle-yaml-in-golang/

func YamlToJson(yamlBytes []byte) ([]byte, error) {
	return yaml.YAMLToJSON(yamlBytes)
}

func YamlUnmarshal(yamlBytes []byte, dest interface{}) error {
	return yaml.Unmarshal(yamlBytes, dest)
}

func YamlMarshal(o interface{}) ([]byte, error) {
	return yaml.Marshal(o)
}
