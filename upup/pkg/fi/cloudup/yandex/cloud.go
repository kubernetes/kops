/*
Copyright 2022 The Kubernetes Authors.

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

package yandex

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	ycsdk "github.com/yandex-cloud/go-sdk"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	TagKubernetesClusterName      = "kops.k8s.io/cluster"
	TagKubernetesFirewallRole     = "kops.k8s.io/firewall-role"
	TagKubernetesInstanceGroup    = "kops.k8s.io/instance-group"
	TagKubernetesInstanceRole     = "kops.k8s.io/instance-role"
	TagKubernetesInstanceUserData = "kops.k8s.io/instance-userdata"
	TagKubernetesVolumeRole       = "kops.k8s.io/volume-role"
)

// YandexCloud exposes all the interfaces required to operate on Yandex Cloud resources
type YandexCloud interface {
	fi.Cloud
	SDK() *ycsdk.SDK
}

// static compile time check to validate YandexCloud's fi.Cloud Interface.
var _ fi.Cloud = &yandexCloudImplementation{}

// yandexCloudImplementation holds the sdk client object to interact with Yandex resources.
type yandexCloudImplementation struct {
	Client *ycsdk.SDK

	dns dnsprovider.Interface

	region string
}

/* s3-kops-iam.json
→ yc iam key create --service-account-name s3-kops --output s3-kops-iam.json
json has everything which is required
*/
var (
	keyID            string
	serviceAccountID string
	keyFile          []byte
)

// NewYandexCloud returns a Cloud
func NewYandexCloud(region string) (YandexCloud, error) {
	credentialFile := os.Getenv("YANDEX_CLOUD_CREDENTIAL_FILE")
	if credentialFile == "" {
		log.Fatal(errors.New("YANDEX_CLOUD_CREDENTIAL_FILE is required"))
	}

	account, err := ioutil.ReadFile(credentialFile)

	if err != nil {
		log.Fatal(err)
	}
	var data struct {
		Id               string `json:"id"`
		ServiceAccountId string `json:"service_account_id"`
		CreatedAt        string `json:"created_at"`
		KeyAlgorithm     string `json:"key_algorithm"`
		PublicKey        string `json:"public_key"`
		PrivateKey       string `json:"private_key"`
	}

	err = json.Unmarshal(account, &data)
	if err != nil {
		log.Fatal(err)
	}
	keyID = data.Id
	serviceAccountID = data.ServiceAccountId
	keyFile = []byte(data.PrivateKey)

	token := getIAMToken()
	//TODO(YuraBeznos): automate folder read/create
	//TODO(YuraBeznos): add multizone

	// auth for service account
	// https://cloud.yandex.ru/docs/iam/operations/iam-token/create-for-sa
	// 1 jwt -> 2 jwt to iam -> auth
	sdk, err := ycsdk.Build(context.TODO(), ycsdk.Config{
		Credentials: ycsdk.NewIAMTokenCredentials(token),
	})
	if err != nil {
		return nil, err
	}

	return &yandexCloudImplementation{
		Client: sdk,
		dns:    nil,
		region: region,
	}, nil
}

// SDK returns an implementation of ycsdk.SDK
func (c *yandexCloudImplementation) SDK() *ycsdk.SDK {
	return c.Client
}

func (c *yandexCloudImplementation) DNS() (dnsprovider.Interface, error) {
	panic("implement me")
}

// ProviderID returns the kOps API identifier for Yandex Cloud
func (c *yandexCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderYandex
}

// Region returns the Yandex Cloud region
func (c *yandexCloudImplementation) Region() string {
	return c.region
}

func (c *yandexCloudImplementation) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return nil, fmt.Errorf("yandex cloud provider does not implement GetCloudGroups at this time")
}

// FindVPCInfo is not implemented
func (c *yandexCloudImplementation) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, errors.New("yandex cloud provider does not implement FindVPCInfo at this time")
}

// FindClusterStatus is not implemented
func (c *yandexCloudImplementation) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	klog.Warning("yandex cloud provider does not implement FindClusterStatus at this time")
	return nil, nil
}

func (c *yandexCloudImplementation) DeleteInstance(i *cloudinstances.CloudInstance) error {
	panic("implement me")
}

func (c *yandexCloudImplementation) DetachInstance(instance *cloudinstances.CloudInstance) error {
	panic("implement me")
}

// DeleteGroup is not implemented
func (c *yandexCloudImplementation) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return fmt.Errorf("yandex cloud provider does not implement DeleteGroup at this time")
}

// DeregisterInstance is not implemented
func (c *yandexCloudImplementation) DeregisterInstance(i *cloudinstances.CloudInstance) error {
	klog.Warning("yandex cloud provider does not implement DeregisterInstance at this time")
	return nil
}

func (c *yandexCloudImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	return nil, nil
}

// TODO(YuraBeznos): move everything related to authentication to own file
// Create JWT.
func signedToken() string {
	claims := jwt.RegisteredClaims{
		Issuer:    serviceAccountID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Audience:  []string{"https://iam.api.cloud.yandex.net/iam/v1/tokens"},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodPS256, claims)
	token.Header["kid"] = keyID

	privateKey := loadPrivateKey()
	signed, err := token.SignedString(privateKey)
	if err != nil {
		panic(err)
	}
	return signed
}

func loadPrivateKey() *rsa.PrivateKey {
	//data, err := ioutil.ReadFile(keyFile)
	//if err != nil {
	//	panic(err)
	//}
	rsaPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyFile)
	if err != nil {
		panic(err)
	}
	return rsaPrivateKey
}

func getIAMToken() string {
	jot := signedToken()
	fmt.Println(jot)
	resp, err := http.Post(
		"https://iam.api.cloud.yandex.net/iam/v1/tokens",
		"application/json",
		strings.NewReader(fmt.Sprintf(`{"jwt":"%s"}`, jot)),
	)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		panic(fmt.Sprintf("%s: %s", resp.Status, body))
	}
	var data struct {
		IAMToken string `json:"iamToken"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		panic(err)
	}
	return data.IAMToken
}
