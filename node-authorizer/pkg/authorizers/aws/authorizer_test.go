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

package aws

import (
	"context"
	"testing"

	"k8s.io/kops/node-authorizer/pkg/server"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/stretchr/testify/assert"
)

func newTestAuthorizer(t *testing.T, config *server.Config) *awsNodeAuthorizer {
	if config == nil {
		config = &server.Config{}
	}
	c := &awsNodeAuthorizer{
		config: config,
		vpcID:  "test",
	}
	if err := GetPublicCertificates(); err != nil {
		t.Errorf("unable to parse public certificates: %s", err)
		t.FailNow()
	}

	return c
}

func TestValidateIdentityDocument(t *testing.T) {
	c := newTestAuthorizer(t, nil)

	request := &Request{
		Document: []byte(`MIAGCSqGSIb3DQEHAqCAMIACAQExCzAJBgUrDgMCGgUAMIAGCSqGSIb3DQEHAaCAJIAEggHUewog
ICJtYXJrZXRwbGFjZVByb2R1Y3RDb2RlcyIgOiBudWxsLAogICJkZXZwYXlQcm9kdWN0Q29kZXMi
IDogbnVsbCwKICAicHJpdmF0ZUlwIiA6ICIxMC4yNTAuMTAxLjE3IiwKICAidmVyc2lvbiIgOiAi
MjAxNy0wOS0zMCIsCiAgImluc3RhbmNlSWQiIDogImktMDJhYTA0MDdmNDAwYmJmY2YiLAogICJi
aWxsaW5nUHJvZHVjdHMiIDogbnVsbCwKICAiaW5zdGFuY2VUeXBlIiA6ICJtNC5sYXJnZSIsCiAg
ImF2YWlsYWJpbGl0eVpvbmUiIDogImV1LXdlc3QtMmIiLAogICJrZXJuZWxJZCIgOiBudWxsLAog
ICJyYW1kaXNrSWQiIDogbnVsbCwKICAiYWNjb3VudElkIiA6ICI2NzA5MzA2NDYxMDMiLAogICJh
cmNoaXRlY3R1cmUiIDogIng4Nl82NCIsCiAgImltYWdlSWQiIDogImFtaS1iNTMwZDFkMiIsCiAg
InBlbmRpbmdUaW1lIiA6ICIyMDE4LTA2LTA5VDEzOjE5OjQzWiIsCiAgInJlZ2lvbiIgOiAiZXUt
d2VzdC0yIgp9AAAAAAAAMYIBFzCCARMCAQEwaTBcMQswCQYDVQQGEwJVUzEZMBcGA1UECBMQV2Fz
aGluZ3RvbiBTdGF0ZTEQMA4GA1UEBxMHU2VhdHRsZTEgMB4GA1UEChMXQW1hem9uIFdlYiBTZXJ2
aWNlcyBMTEMCCQCWukjZ5V4aZzAJBgUrDgMCGgUAoF0wGAYJKoZIhvcNAQkDMQsGCSqGSIb3DQEH
ATAcBgkqhkiG9w0BCQUxDxcNMTgwNjA5MTMxOTQ3WjAjBgkqhkiG9w0BCQQxFgQUzenf7yQR02cW
A4t1ZTpGNjz7490wCQYHKoZIzjgEAwQuMCwCFAZ1VFJSd81PnuG6+1sDFOgr/3tyAhRH14cLYMTN
Uce4CDpektlneHOeLQAAAAAAAA==`),
	}

	identity := &ec2metadata.EC2InstanceIdentityDocument{}
	reason, err := c.validateIdentityDocument(context.TODO(), request.Document, identity)
	assert.NoError(t, err)
	assert.NotNil(t, identity)
	assert.Empty(t, reason)
	assert.Equal(t, "i-02aa0407f400bbfcf", identity.InstanceID)
}
