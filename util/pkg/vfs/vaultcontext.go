/*
Copyright 2020 The Kubernetes Authors.

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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"k8s.io/klog/v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	vault "github.com/hashicorp/vault/api"
)

func newVaultClient(scheme string, host string, port string) (*vault.Client, error) {
	addr := scheme + host
	if port != "" {
		addr = addr + ":" + port
	}
	config := &vault.Config{
		Address:    addr,
		HttpClient: &http.Client{},
	}
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}

	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		token, err = awsAuth(client, host)
		if err != nil {
			return nil, fmt.Errorf("error authenticating vault to AWS: %v", err)
		}
	}

	client.SetToken(token)
	return client, nil
}

func awsAuth(client *vault.Client, host string) (string, error) {
	klog.Infof("Using AWS IAM to authenticate to Vault at %q", host)
	config := aws.NewConfig()
	config = config.WithCredentialsChainVerboseErrors(true)
	session, err := session.NewSession(config)
	if err != nil {
		return "", err
	}
	svc := sts.New(session)
	stsRequest, _ := svc.GetCallerIdentityRequest(nil)
	stsRequest.HTTPRequest.Header.Add("X-Vault-AWS-IAM-Server-ID", host)
	stsRequest.Sign()
	headersJson, err := json.Marshal(stsRequest.HTTPRequest.Header)
	if err != nil {
		return "", err
	}
	requestBody, err := ioutil.ReadAll(stsRequest.HTTPRequest.Body)
	if err != nil {
		return "", err
	}
	loginData := make(map[string]interface{})
	loginData["iam_http_request_method"] = stsRequest.HTTPRequest.Method
	loginData["iam_request_url"] = base64.StdEncoding.EncodeToString([]byte(stsRequest.HTTPRequest.URL.String()))
	loginData["iam_request_headers"] = base64.StdEncoding.EncodeToString(headersJson)
	loginData["iam_request_body"] = base64.StdEncoding.EncodeToString(requestBody)
	loginData["role"] = ""
	path := "auth/aws/login"
	secret, err := client.Logical().Write(path, loginData)
	if err != nil {
		return "", err
	}
	return secret.Auth.ClientToken, err
}
