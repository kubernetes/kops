// +build fixtures

package adminactions

import (
	"net/http"
	"testing"

	th "github.com/rackspace/gophercloud/testhelper"
	"github.com/rackspace/gophercloud/testhelper/client"
)

func mockCreateBackupResponse(t *testing.T, id string) {
	th.Mux.HandleFunc("/servers/"+id+"/action", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "POST")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)
		th.TestJSONRequest(t, r, `{"createBackup": {"name": "Backup 1", "backup_type": "daily", "rotation": 1}}`)
		w.WriteHeader(http.StatusAccepted)
	})
}

func mockInjectNetworkInfoResponse(t *testing.T, id string) {
	th.Mux.HandleFunc("/servers/"+id+"/action", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "POST")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)
		th.TestJSONRequest(t, r, `{"injectNetworkInfo": ""}`)
		w.WriteHeader(http.StatusAccepted)
	})
}

func mockMigrateResponse(t *testing.T, id string) {
	th.Mux.HandleFunc("/servers/"+id+"/action", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "POST")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)
		th.TestJSONRequest(t, r, `{"migrate": ""}`)
		w.WriteHeader(http.StatusAccepted)
	})
}

const liveMigrateRequest = `{"os-migrateLive": {"host": "", "disk_over_commit": false, "block_migration": true}}`
const targetLiveMigrateRequest = `{"os-migrateLive": {"host": "target-compute", "disk_over_commit": false, "block_migration": true}}`

func mockLiveMigrateResponse(t *testing.T, id string) {
	th.Mux.HandleFunc("/servers/"+id+"/action", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "POST")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)
		th.TestJSONRequest(t, r, liveMigrateRequest)
		w.WriteHeader(http.StatusAccepted)
	})
}

func mockTargetLiveMigrateResponse(t *testing.T, id string) {
	th.Mux.HandleFunc("/servers/"+id+"/action", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "POST")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)
		th.TestJSONRequest(t, r, targetLiveMigrateRequest)
		w.WriteHeader(http.StatusAccepted)
	})
}

func mockResetNetworkResponse(t *testing.T, id string) {
	th.Mux.HandleFunc("/servers/"+id+"/action", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "POST")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)
		th.TestJSONRequest(t, r, `{"resetNetwork": ""}`)
		w.WriteHeader(http.StatusAccepted)
	})
}

func mockResetStateResponse(t *testing.T, id string) {
	th.Mux.HandleFunc("/servers/"+id+"/action", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "POST")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)
		th.TestJSONRequest(t, r, `{"os-resetState": {"state": "active"}}`)
		w.WriteHeader(http.StatusAccepted)
	})
}
