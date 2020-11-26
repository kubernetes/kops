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

package controllers

import (
	"testing"

	"k8s.io/kops/pkg/pki"

	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
)

func Test_SignCSR(t *testing.T) {
	csr := &cmapi.CertificateRequest{
		Spec: cmapi.CertificateRequestSpec{
			Request: []byte("-----BEGIN CERTIFICATE REQUEST-----\nMIIC6DCCAdACAQAwLzEUMBIGA1UEChMLa29wcy5rOHMuaW8xFzAVBgNVBAMTDm1l\ndHJpY3Mtc2VydmVyMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxaKv\nKZud7+qiH2Dwf4L9F/2d8sj5GlX3G0MkjxDquHZ2jnPGTsczqCbI5VBKtKes+sjL\nh5WWpsYIRke7w/sxKox/pkPnR4ldBYBry51vhg94IjFQmoMXy51N0lly5dHN9T88\nQO5wtCyTO7A3bn8S23FDklS8V6NT56bS5Wm76jGukzkQVxT7BNJbpS0SP2NqrOAL\nUcaelGwnpxs80QzaiBOOG0vv+25Wvencc/ryD12EWPPqdRSDx4eBUjpTb7wY+Ify\nj6FTiqevlVYKJOj2hXHtY0+SZ+g+KDAo6HAplgwfizXvBzHQOx9iVoXgla0VkRA5\npk25Gv/Xq3QQ5Ql6fwIDAQABoHQwcgYJKoZIhvcNAQkOMWUwYzA1BgNVHREELjAs\ngg5tZXRyaWNzLXNlcnZlcoIabWV0cmljcy1zZXJ2ZXIua3ViZS1zeXN0ZW0wCwYD\nVR0PBAQDAgAAMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjANBgkqhkiG\n9w0BAQsFAAOCAQEAFzaRTwWsTNEty60t4l+Iq6JoX3t+fBUn/QQEqYpD/Vbv6z4z\ngW80a4Cg6QHQMv1dHQ7l4aEn4gzvDqkayGBFc8f0au1qU1XeYxA1lKKkpj7yIpV5\nwq1YSWFhUG7frTGjD3huPs10noBYR4e+aUgT8UGPG649ISZ8FXVzVkaSBt1hGeEo\nRLi56n5POQHB/wL2GOuajtsIipk5acK2iHtxCubPiIKv6UXEdgNfMuv5LTmr5aqK\nCVCte7Y1i97yjQUDjyh22lCkiXqik24J7zmK9s3PeJ4W44O3YZ25KXqT6PGpAh/m\neA1/T2/HAXoU7Ou0LSTY3MDsAedS0BngqgJ6ig==\n-----END CERTIFICATE REQUEST-----\n"),
			Usages: []cmapi.KeyUsage{
				cmapi.UsageServerAuth,
				cmapi.UsageClientAuth,
			},
		},
	}

	keystore, _ := pki.NewMockKeystore()

	err := signCSR(csr, keystore)
	if err != nil {
		t.Fatalf("could not sign CSR: %v", err)
	}
}
