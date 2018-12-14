package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var certJSONResponse = `
{
	"certificate": {
		"id": "892071a0-bb95-49bc-8021-3afd67a210bf",
		"name": "web-cert-01",
		"not_after": "2017-02-22T00:23:00Z",
		"sha1_fingerprint": "dfcc9f57d86bf58e321c2c6c31c7a971be244ac7",
		"created_at": "2017-02-08T16:02:37Z"
	}
}
`

var certsJSONResponse = `
{
  	"certificates": [
    	{
      		"id": "892071a0-bb95-49bc-8021-3afd67a210bf",
      		"name": "web-cert-01",
      		"not_after": "2017-02-22T00:23:00Z",
      		"sha1_fingerprint": "dfcc9f57d86bf58e321c2c6c31c7a971be244ac7",
      		"created_at": "2017-02-08T16:02:37Z"
    	},
    	{
      		"id": "992071a0-bb95-49bc-8021-3afd67a210bf",
      		"name": "web-cert-02",
      		"not_after": "2017-02-22T00:23:00Z",
      		"sha1_fingerprint": "cfcc9f57d86bf58e321c2c6c31c7a971be244ac7",
      		"created_at": "2017-02-08T16:02:37Z"
    	}
  	],
  	"links": {},
  	"meta": {
    	"total": 1
  	}
}
`

func TestCertificates_Get(t *testing.T) {
	setup()
	defer teardown()

	urlStr := "/v2/certificates"
	cID := "892071a0-bb95-49bc-8021-3afd67a210bf"
	urlStr = path.Join(urlStr, cID)
	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, certJSONResponse)
	})

	certificate, _, err := client.Certificates.Get(ctx, cID)
	if err != nil {
		t.Errorf("Certificates.Get returned error: %v", err)
	}

	expected := &Certificate{
		ID:              "892071a0-bb95-49bc-8021-3afd67a210bf",
		Name:            "web-cert-01",
		NotAfter:        "2017-02-22T00:23:00Z",
		SHA1Fingerprint: "dfcc9f57d86bf58e321c2c6c31c7a971be244ac7",
		Created:         "2017-02-08T16:02:37Z",
	}

	assert.Equal(t, expected, certificate)
}

func TestCertificates_List(t *testing.T) {
	setup()
	defer teardown()

	urlStr := "/v2/certificates"
	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, certsJSONResponse)
	})

	certificates, _, err := client.Certificates.List(ctx, nil)

	if err != nil {
		t.Errorf("Certificates.List returned error: %v", err)
	}

	expected := []Certificate{
		{
			ID:              "892071a0-bb95-49bc-8021-3afd67a210bf",
			Name:            "web-cert-01",
			NotAfter:        "2017-02-22T00:23:00Z",
			SHA1Fingerprint: "dfcc9f57d86bf58e321c2c6c31c7a971be244ac7",
			Created:         "2017-02-08T16:02:37Z",
		},
		{
			ID:              "992071a0-bb95-49bc-8021-3afd67a210bf",
			Name:            "web-cert-02",
			NotAfter:        "2017-02-22T00:23:00Z",
			SHA1Fingerprint: "cfcc9f57d86bf58e321c2c6c31c7a971be244ac7",
			Created:         "2017-02-08T16:02:37Z",
		},
	}

	assert.Equal(t, expected, certificates)
}

func TestCertificates_Create(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &CertificateRequest{
		Name:             "web-cert-01",
		PrivateKey:       "-----BEGIN PRIVATE KEY-----",
		LeafCertificate:  "-----BEGIN CERTIFICATE-----",
		CertificateChain: "-----BEGIN CERTIFICATE-----",
	}

	urlStr := "/v2/certificates"
	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		v := new(CertificateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		assert.Equal(t, createRequest, v)

		fmt.Fprint(w, certJSONResponse)
	})

	certificate, _, err := client.Certificates.Create(ctx, createRequest)
	if err != nil {
		t.Errorf("Certificates.Create returned error: %v", err)
	}

	expected := &Certificate{
		ID:              "892071a0-bb95-49bc-8021-3afd67a210bf",
		Name:            "web-cert-01",
		NotAfter:        "2017-02-22T00:23:00Z",
		SHA1Fingerprint: "dfcc9f57d86bf58e321c2c6c31c7a971be244ac7",
		Created:         "2017-02-08T16:02:37Z",
	}

	assert.Equal(t, expected, certificate)
}

func TestCertificates_Delete(t *testing.T) {
	setup()
	defer teardown()

	cID := "892071a0-bb95-49bc-8021-3afd67a210bf"
	urlStr := "/v2/certificates"
	urlStr = path.Join(urlStr, cID)
	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Certificates.Delete(ctx, cID)

	if err != nil {
		t.Errorf("Certificates.Delete returned error: %v", err)
	}
}
