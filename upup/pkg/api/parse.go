package api

import (
	"fmt"
	"k8s.io/kops/upup/pkg/fi/utils"
	"strings"
)

func ParseYaml(data []byte, dest interface{}) error {
	// Yaml can't parse empty strings
	configString := string(data)
	configString = strings.TrimSpace(configString)

	if configString != "" {
		err := utils.YamlUnmarshal([]byte(configString), dest)
		if err != nil {
			return fmt.Errorf("error parsing configuration: %v", err)
		}
	}

	return nil
}

func ToYaml(dest interface{}) ([]byte, error) {
	data, err := utils.YamlMarshal(dest)
	if err != nil {
		return nil, fmt.Errorf("error converting to yaml: %v", err)
	}

	return data, nil
}
