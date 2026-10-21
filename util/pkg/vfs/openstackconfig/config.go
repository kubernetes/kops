/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package openstackconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-ini/ini"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

type OpenstackConfig struct{}

func (OpenstackConfig) filename() (string, error) {
	name := os.Getenv("OPENSTACK_CREDENTIAL_FILE")
	if name != "" {
		klog.V(2).Infof("using openstack config found in $OPENSTACK_CREDENTIAL_FILE: %s", name)
		return name, nil
	}

	homeDir := homedir.HomeDir()
	if homeDir == "" {
		return "", fmt.Errorf("can not find home directory")
	}
	f := filepath.Join(homeDir, ".openstack", "config")
	klog.V(2).Infof("using openstack config found in %s", f)
	return f, nil
}

func (oc OpenstackConfig) getSection(name string, items []string) (map[string]string, error) {
	filename, err := oc.filename()
	if err != nil {
		return nil, err
	}
	config, err := ini.Load(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading config file: %v", err)
	}
	section, err := config.GetSection(name)
	if err != nil {
		return nil, fmt.Errorf("error getting section of %s: %v", name, err)
	}
	values := make(map[string]string)
	for _, item := range items {
		values[item] = section.Key(item).String()
	}
	return values, nil
}

func (oc OpenstackConfig) GetCredential() (gophercloud.AuthOptions, error) {
	// prioritize environment config
	env, enverr := openstack.AuthOptionsFromEnv()
	if enverr != nil {
		klog.Warningf("Could not initialize OpenStack config from environment: %v", enverr)
		// fallback to config file
		return oc.getCredentialFromFile()
	}

	if env.ApplicationCredentialID != "" && env.Username == "" {
		env.Scope = &gophercloud.AuthScope{}
	}
	env.AllowReauth = true
	return env, nil
}

func (oc OpenstackConfig) GetInsecureSkipVerify() bool {
	s := os.Getenv(EnvKeyOpenstackTLSInsecureSkipVerify)
	if s == "true" || s == "1" {
		return true
	}
	return false
}

func (oc OpenstackConfig) GetRegion() (string, error) {
	var region string
	if region = os.Getenv("OS_REGION_NAME"); region != "" {
		if len(region) > 1 {
			if region[0] == '\'' && region[len(region)-1] == '\'' {
				region = region[1 : len(region)-1]
			}
		}
		return region, nil
	}

	items := []string{"region"}
	// TODO: Unsure if this is the correct section for region
	values, err := oc.getSection("Global", items)
	if err != nil {
		return "", fmt.Errorf("region not provided in OS_REGION_NAME or openstack config section GLOBAL")
	}
	return values["region"], nil
}

func (oc OpenstackConfig) getCredentialFromFile() (gophercloud.AuthOptions, error) {
	opt := gophercloud.AuthOptions{}
	name := "Default"
	items := []string{"identity", "user", "user_id", "password", "domain_id", "domain_name", "tenant_id", "tenant_name"}
	values, err := oc.getSection(name, items)
	if err != nil {
		return opt, err
	}

	for _, c1 := range []string{"identity", "password"} {
		if values[c1] == "" {
			return opt, fmt.Errorf("missing %s in section of %s", c1, name)
		}
	}

	checkItems := [][]string{{"user", "user_id"}, {"domain_name", "domain_id"}, {"tenant_name", "tenant_id"}}
	for _, c2 := range checkItems {
		if values[c2[0]] == "" && values[c2[1]] == "" {
			return opt, fmt.Errorf("missing %s and %s in section of %s", c2[0], c2[1], name)
		}
	}

	opt.IdentityEndpoint = values["identity"]
	opt.UserID = values["user_id"]
	opt.Username = values["user"]
	opt.Password = values["password"]
	opt.TenantID = values["tenant_id"]
	opt.TenantName = values["tenant_name"]
	opt.DomainID = values["domain_id"]
	opt.DomainName = values["domain_name"]
	opt.AllowReauth = true

	return opt, nil
}

func (oc OpenstackConfig) GetServiceConfig(name string) (gophercloud.EndpointOpts, error) {
	opt := gophercloud.EndpointOpts{}
	items := []string{"service_type", "service_name", "region", "availability"}
	values, err := oc.getSection(name, items)
	if err != nil {
		return opt, err
	}

	if values["region"] == "" {
		return opt, fmt.Errorf("missing region in section of %s", name)
	}

	opt.Type = values["service_type"]
	opt.Name = values["service_name"]
	opt.Region = values["region"]
	opt.Availability = gophercloud.Availability(values["availability"])

	return opt, nil
}
