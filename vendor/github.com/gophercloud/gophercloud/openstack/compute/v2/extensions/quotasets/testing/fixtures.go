package testing

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/quotasets"
	th "github.com/gophercloud/gophercloud/testhelper"
	"github.com/gophercloud/gophercloud/testhelper/client"
)

// GetOutput is a sample response to a Get call.
const GetOutput = `
{
   "quota_set" : {
      "instances" : 25,
      "security_groups" : 10,
      "security_group_rules" : 20,
      "cores" : 200,
      "injected_file_content_bytes" : 10240,
      "injected_files" : 5,
      "metadata_items" : 128,
      "ram" : 200000,
      "key_pairs" : 10,
      "injected_file_path_bytes" : 255,
	  "server_groups" : 2,
	  "server_group_members" : 3
   }
}
`
const FirstTenantID = "555544443333222211110000ffffeeee"

// FirstQuotaSet is the first result in ListOutput.
var FirstQuotaSet = quotasets.QuotaSet{
	FixedIps:                 0,
	FloatingIps:              0,
	InjectedFileContentBytes: 10240,
	InjectedFilePathBytes:    255,
	InjectedFiles:            5,
	KeyPairs:                 10,
	MetadataItems:            128,
	Ram:                      200000,
	SecurityGroupRules:       20,
	SecurityGroups:           10,
	Cores:                    200,
	Instances:                25,
	ServerGroups:             2,
	ServerGroupMembers:       3,
}

//The expected update Body. Is also returned by PUT request
const UpdateOutput = `{"quota_set":{"cores":200,"fixed_ips":0,"floating_ips":0,"injected_file_content_bytes":10240,"injected_file_path_bytes":255,"injected_files":5,"instances":25,"key_pairs":10,"metadata_items":128,"ram":200000,"security_group_rules":20,"security_groups":10,"server_groups":2,"server_group_members":3}}`

//The expected partialupdate Body. Is also returned by PUT request
const PartialUpdateBody = `{"quota_set":{"cores":200, "force":true}}`

//Result of Quota-update
var UpdatedQuotaSet = quotasets.UpdateOpts{
	FixedIps:                 gophercloud.IntToPointer(0),
	FloatingIps:              gophercloud.IntToPointer(0),
	InjectedFileContentBytes: gophercloud.IntToPointer(10240),
	InjectedFilePathBytes:    gophercloud.IntToPointer(255),
	InjectedFiles:            gophercloud.IntToPointer(5),
	KeyPairs:                 gophercloud.IntToPointer(10),
	MetadataItems:            gophercloud.IntToPointer(128),
	Ram:                      gophercloud.IntToPointer(200000),
	SecurityGroupRules:       gophercloud.IntToPointer(20),
	SecurityGroups:           gophercloud.IntToPointer(10),
	Cores:                    gophercloud.IntToPointer(200),
	Instances:                gophercloud.IntToPointer(25),
	ServerGroups:             gophercloud.IntToPointer(2),
	ServerGroupMembers:       gophercloud.IntToPointer(3),
}

// HandleGetSuccessfully configures the test server to respond to a Get request for sample tenant
func HandleGetSuccessfully(t *testing.T) {
	th.Mux.HandleFunc("/os-quota-sets/"+FirstTenantID, func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)

		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, GetOutput)
	})
}

// HandlePutSuccessfully configures the test server to respond to a Put request for sample tenant
func HandlePutSuccessfully(t *testing.T) {
	th.Mux.HandleFunc("/os-quota-sets/"+FirstTenantID, func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "PUT")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)
		th.TestJSONRequest(t, r, UpdateOutput)
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, UpdateOutput)
	})
}

// HandlePartialPutSuccessfully configures the test server to respond to a Put request for sample tenant that only containes specific values
func HandlePartialPutSuccessfully(t *testing.T) {
	th.Mux.HandleFunc("/os-quota-sets/"+FirstTenantID, func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "PUT")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)
		th.TestJSONRequest(t, r, PartialUpdateBody)
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, UpdateOutput)
	})
}

// HandleDeleteSuccessfully configures the test server to respond to a Delete request for sample tenant
func HandleDeleteSuccessfully(t *testing.T) {
	th.Mux.HandleFunc("/os-quota-sets/"+FirstTenantID, func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "DELETE")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)
		th.TestBody(t, r, "")
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(202)
	})
}
