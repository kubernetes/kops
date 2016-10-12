package registry

import (
	"fmt"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/vfs"
	"os"
	"strings"
)

func ReadConfig(configPath vfs.Path, config interface{}) error {
	data, err := configPath.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("error reading configuration file %s: %v", configPath, err)
	}

	// Yaml can't parse empty strings
	configString := string(data)
	configString = strings.TrimSpace(configString)

	if configString != "" {
		err = utils.YamlUnmarshal([]byte(configString), config)
		if err != nil {
			return fmt.Errorf("error parsing configuration: %v", err)
		}
	}

	return nil
}

func WriteConfig(configPath vfs.Path, config interface{}, writeOptions ...vfs.WriteOption) error {
	data, err := utils.YamlMarshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling configuration: %v", err)
	}

	create := false
	for _, writeOption := range writeOptions {
		switch writeOption {
		case vfs.WriteOptionCreate:
			create = true
		case vfs.WriteOptionOnlyIfExists:
			_, err = configPath.ReadFile()
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("cannot update configuration file %s: does not exist", configPath)
				}
				return fmt.Errorf("error checking if configuration file %s exists already: %v", configPath, err)
			}
		default:
			return fmt.Errorf("unknown write option: %q", writeOption)
		}
	}

	if create {
		err = configPath.CreateFile(data)
	} else {
		err = configPath.WriteFile(data)
	}
	if err != nil {
		return fmt.Errorf("error writing configuration file %s: %v", configPath, err)
	}
	return nil
}
