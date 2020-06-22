/*
Copyright 2019 The Kubernetes Authors.

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

package vfs

import (
	"fmt"
	"os"
	"strings"

	"github.com/denverdino/aliyungo/metadata"
	"github.com/denverdino/aliyungo/oss"
)

type aliyunOSSConfig struct {
	region          oss.Region
	internal        bool
	accessKeyID     string
	accessKeySecret string
	securityToken   string
	secure          bool
}

func NewOSSPath(client *oss.Client, bucket string, key string) (*OSSPath, error) {
	bucket = strings.TrimSuffix(bucket, "/")
	key = strings.TrimPrefix(key, "/")

	return &OSSPath{
		client: client,
		bucket: bucket,
		key:    key,
	}, nil
}

func NewAliOSSClient() (*oss.Client, error) {
	c := &aliyunOSSConfig{}
	err := c.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("error building aliyun oss client: %v", err)
	}

	if c.securityToken != "" {
		return oss.NewOSSClientForAssumeRole(c.region, c.internal, c.accessKeyID, c.accessKeySecret, c.securityToken, c.secure), nil
	}

	return oss.NewOSSClient(c.region, c.internal, c.accessKeyID, c.accessKeySecret, c.secure), nil
}

func (c *aliyunOSSConfig) loadConfig() error {
	meta := metadata.NewMetaData(nil)

	c.region = oss.Region(os.Getenv("OSS_REGION"))
	if c.region == "" {
		region, err := meta.Region()
		if err != nil {
			return fmt.Errorf("can't get region-id from ECS metadata")
		}
		c.region = oss.Region(fmt.Sprintf("oss-%s", region))
	}

	c.accessKeyID = os.Getenv("ALIYUN_ACCESS_KEY_ID")
	if c.accessKeyID != "" {
		c.accessKeySecret = os.Getenv("ALIYUN_ACCESS_KEY_SECRET")
		if c.accessKeySecret == "" {
			return fmt.Errorf("ALIYUN_ACCESS_KEY_SECRET cannot be empty")
		}
	} else {
		role, err := meta.RoleName()
		if err != nil {
			return fmt.Errorf("Can't find role from ECS metadata: %s", err)
		}

		roleAuth, err := meta.RamRoleToken(role)
		if err != nil {
			return fmt.Errorf("Can't get role token: %s", err)
		}
		c.accessKeyID = roleAuth.AccessKeyId
		c.accessKeySecret = roleAuth.AccessKeySecret
		c.securityToken = roleAuth.SecurityToken
	}

	ossInternal := os.Getenv("ALIYUN_OSS_INTERNAL")
	if ossInternal != "" {
		c.internal = true
	} else {
		c.internal = false
	}
	c.secure = true
	return nil
}
