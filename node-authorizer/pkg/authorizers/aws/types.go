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

// Request is the request the node authorizer
type Request struct {
	// Document is the PKCS7 signed identity document
	Document []byte
}

var (
	// awsCertificates is a collection of AWS public certificates used to sign the identity documents
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-identity-documents.html
	awsCertificates = []string{
		// AWS Public Certificate
		`-----BEGIN CERTIFICATE-----
MIIC7TCCAq0CCQCWukjZ5V4aZzAJBgcqhkjOOAQDMFwxCzAJBgNVBAYTAlVTMRkw
FwYDVQQIExBXYXNoaW5ndG9uIFN0YXRlMRAwDgYDVQQHEwdTZWF0dGxlMSAwHgYD
VQQKExdBbWF6b24gV2ViIFNlcnZpY2VzIExMQzAeFw0xMjAxMDUxMjU2MTJaFw0z
ODAxMDUxMjU2MTJaMFwxCzAJBgNVBAYTAlVTMRkwFwYDVQQIExBXYXNoaW5ndG9u
IFN0YXRlMRAwDgYDVQQHEwdTZWF0dGxlMSAwHgYDVQQKExdBbWF6b24gV2ViIFNl
cnZpY2VzIExMQzCCAbcwggEsBgcqhkjOOAQBMIIBHwKBgQCjkvcS2bb1VQ4yt/5e
ih5OO6kK/n1Lzllr7D8ZwtQP8fOEpp5E2ng+D6Ud1Z1gYipr58Kj3nssSNpI6bX3
VyIQzK7wLclnd/YozqNNmgIyZecN7EglK9ITHJLP+x8FtUpt3QbyYXJdmVMegN6P
hviYt5JH/nYl4hh3Pa1HJdskgQIVALVJ3ER11+Ko4tP6nwvHwh6+ERYRAoGBAI1j
k+tkqMVHuAFcvAGKocTgsjJem6/5qomzJuKDmbJNu9Qxw3rAotXau8Qe+MBcJl/U
hhy1KHVpCGl9fueQ2s6IL0CaO/buycU1CiYQk40KNHCcHfNiZbdlx1E9rpUp7bnF
lRa2v1ntMX3caRVDdbtPEWmdxSCYsYFDk4mZrOLBA4GEAAKBgEbmeve5f8LIE/Gf
MNmP9CM5eovQOGx5ho8WqD+aTebs+k2tn92BBPqeZqpWRa5P/+jrdKml1qx4llHW
MXrs3IgIb6+hUIB+S8dz8/mmO0bpr76RoZVCXYab2CZedFut7qc3WUH9+EUAH5mw
vSeDCOUMYQR7R9LINYwouHIziqQYMAkGByqGSM44BAMDLwAwLAIUWXBlk40xTwSw
7HX32MxXYruse9ACFBNGmdX2ZBrVNGrN9N2f6ROk0k9K
-----END CERTIFICATE-----`,
		// AWS GovCloud (US) region
		`-----BEGIN CERTIFICATE-----
MIICuzCCAiQCCQDrSGnlRgvSazANBgkqhkiG9w0BAQUFADCBoTELMAkGA1UEBhMC
VVMxCzAJBgNVBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMRMwEQYDVQQKEwpBbWF6
b24uY29tMRYwFAYDVQQLEw1FQzIgQXV0aG9yaXR5MRowGAYDVQQDExFFQzIgQU1J
IEF1dGhvcml0eTEqMCgGCSqGSIb3DQEJARYbZWMyLWluc3RhbmNlLWlpZEBhbWF6
b24uY29tMB4XDTExMDgxMjE3MTgwNVoXDTIxMDgwOTE3MTgwNVowgaExCzAJBgNV
BAYTAlVTMQswCQYDVQQIEwJXQTEQMA4GA1UEBxMHU2VhdHRsZTETMBEGA1UEChMK
QW1hem9uLmNvbTEWMBQGA1UECxMNRUMyIEF1dGhvcml0eTEaMBgGA1UEAxMRRUMy
IEFNSSBBdXRob3JpdHkxKjAoBgkqhkiG9w0BCQEWG2VjMi1pbnN0YW5jZS1paWRA
YW1hem9uLmNvbTCBnzANBgkqhkiG9w0BAQEFAAOBjQAwgYkCgYEAqaIcGFFTx/SO
1W5G91jHvyQdGP25n1Y91aXCuOOWAUTvSvNGpXrI4AXNrQF+CmIOC4beBASnHCx0
82jYudWBBl9Wiza0psYc9flrczSzVLMmN8w/c78F/95NfiQdnUQPpvgqcMeJo82c
gHkLR7XoFWgMrZJqrcUK0gnsQcb6kakCAwEAATANBgkqhkiG9w0BAQUFAAOBgQDF
VH0+UGZr1LCQ78PbBH0GreiDqMFfa+W8xASDYUZrMvY3kcIelkoIazvi4VtPO7Qc
yAiLr6nkk69Tr/MITnmmsZJZPetshqBndRyL+DaTRnF0/xvBQXj5tEh+AmRjvGtp
6iS1rQoNanN8oEcT2j4b48rmCmnDhRoBcFHwCYs/3w==
-----END CERTIFICATE-----`,
	}
)
