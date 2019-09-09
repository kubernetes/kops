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

	"github.com/denverdino/aliyungo/oss"
)

type aliyunOSSConfig struct {
	region          oss.Region
	internal        bool
	accessKeyId     string
	accessKeySecret string
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

	return oss.NewOSSClient(c.region, c.internal, c.accessKeyId, c.accessKeySecret, c.secure), nil
}

func (c *aliyunOSSConfig) loadConfig() error {
	c.region = oss.Region(os.Getenv("OSS_REGION"))
	if c.region == "" {
		// TODO: can we use default region?
		return fmt.Errorf("OSS_REGION cannot be empty")
	}
	c.accessKeyId = os.Getenv("ALIYUN_ACCESS_KEY_ID")
	if c.accessKeyId == "" {
		return fmt.Errorf("ALIYUN_ACCESS_KEY_ID cannot be empty")
	}
	c.accessKeySecret = os.Getenv("ALIYUN_ACCESS_KEY_SECRET")
	if c.accessKeySecret == "" {
		return fmt.Errorf("ALIYUN_ACCESS_KEY_SECRET cannot be empty")
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
