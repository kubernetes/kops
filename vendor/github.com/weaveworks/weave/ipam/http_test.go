package ipam

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave/common"
	"github.com/weaveworks/weave/net/address"
)

func HTTPPost(t *testing.T, url string) string {
	resp, err := http.Post(url, "", nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "http response")
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

func HTTPGet(t *testing.T, url string) string {
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

func doHTTP(method string, url string) (resp *http.Response, err error) {
	req, _ := http.NewRequest(method, url, nil)
	return http.DefaultClient.Do(req)
}

func listenHTTP(alloc *Allocator, subnet address.CIDR) int {
	router := mux.NewRouter()
	router.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, fmt.Sprintln(alloc))
	})
	alloc.HandleHTTP(router, subnet, "", nil)

	httpListener, err := net.Listen("tcp", ":0")
	if err != nil {
		common.Log.Fatal("Unable to create http listener: ", err)
	}

	go func() {
		srv := &http.Server{Handler: router}
		if err := srv.Serve(httpListener); err != nil {
			common.Log.Fatal("Unable to serve http: ", err)
		}
	}()
	return httpListener.Addr().(*net.TCPAddr).Port
}

func identURL(port int, containerID string) string {
	return fmt.Sprintf("http://localhost:%d/ip/%s", port, containerID)
}

func allocURL(port int, cidr string, containerID string) string {
	return fmt.Sprintf("http://localhost:%d/ip/%s/%s", port, containerID, cidr)
}

func TestHttp(t *testing.T) {
	var (
		containerID = "deadbeef"
		container2  = "baddf00d"
		container3  = "b01df00d"
		universe    = "10.0.0.0/8"
		testCIDR1   = "10.0.3.8/29"
		testCIDR2   = "10.2.0.0/16"
		testAddr1   = "10.0.3.9/29"
		testAddr2   = "10.2.0.1/16"
	)

	alloc, _ := makeAllocatorWithMockGossip(t, "08:00:27:01:c3:9a", universe, 1)
	cidr, _ := address.ParseCIDR(universe)
	port := listenHTTP(alloc, cidr)
	alloc.claimRingForTesting()

	// Allocate an address in each subnet, and check we got what we expected
	cidr1a := HTTPPost(t, allocURL(port, testCIDR1, containerID))
	require.Equal(t, testAddr1, cidr1a, "address")
	cidr2a := HTTPPost(t, allocURL(port, testCIDR2, containerID))
	require.Equal(t, testAddr2, cidr2a, "address")
	// Now, make the same requests again to check the operation is idempotent
	check := HTTPGet(t, allocURL(port, testCIDR1, containerID))
	require.Equal(t, cidr1a, check, "address")
	check = HTTPGet(t, allocURL(port, testCIDR2, containerID))
	require.Equal(t, cidr2a, check, "address")

	// Ask the http server for a pair of addresses for another container and check they're different
	cidr1b := HTTPPost(t, allocURL(port, testCIDR1, container2))
	require.False(t, cidr1b == testAddr1, "address")
	cidr2b := HTTPPost(t, allocURL(port, testCIDR2, container2))
	require.False(t, cidr2b == testAddr2, "address")

	// Now free the first container, and we should get its addresses back when we ask
	doHTTP("DELETE", identURL(port, containerID))

	cidr1c := HTTPPost(t, allocURL(port, testCIDR1, container3))
	require.Equal(t, testAddr1, cidr1c, "address")
	cidr2c := HTTPPost(t, allocURL(port, testCIDR2, container3))
	require.Equal(t, testAddr2, cidr2c, "address")

	// Would like to shut down the http server at the end of this test
	// but it's complicated.
	// See https://groups.google.com/forum/#!topic/golang-nuts/vLHWa5sHnCE
}

func TestBadHttp(t *testing.T) {
	var (
		containerID = "deadbeef"
		testCIDR1   = "10.0.0.0/8"
	)

	alloc, _ := makeAllocatorWithMockGossip(t, "08:00:27:01:c3:9a", testCIDR1, 1)
	defer alloc.Stop()
	cidr, _ := address.ParseCIDR(testCIDR1)
	port := listenHTTP(alloc, cidr)

	alloc.claimRingForTesting()
	cidr1 := HTTPPost(t, allocURL(port, testCIDR1, containerID))
	parts := strings.Split(cidr1, "/")
	testAddr1 := parts[0]
	// Verb that's not handled
	resp, err := doHTTP("HEAD", fmt.Sprintf("http://localhost:%d/ip/%s/%s", port, containerID, testAddr1))
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode, "http response")
	// Mis-spelled URL
	resp, err = doHTTP("POST", fmt.Sprintf("http://localhost:%d/xip/%s/", port, containerID))
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode, "http response")
	// Malformed URL
	resp, err = doHTTP("POST", fmt.Sprintf("http://localhost:%d/ip/%s/foo/bar/baz", port, containerID))
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode, "http response")
}

func TestHTTPCancel(t *testing.T) {
	var (
		containerID = "deadbeef"
		testCIDR1   = "10.0.3.0/29"
	)

	// Say quorum=2, so the allocate won't go ahead
	alloc, _ := makeAllocatorWithMockGossip(t, "08:00:27:01:c3:9a", testCIDR1, 2)
	defer alloc.Stop()
	ExpectBroadcastMessage(alloc, nil) // trying to form consensus
	cidr, _ := address.ParseCIDR(testCIDR1)
	port := listenHTTP(alloc, cidr)

	// Ask the http server for a new address
	req, _ := http.NewRequest("POST", allocURL(port, testCIDR1, containerID), nil)
	// On another goroutine, wait for a bit then cancel the request
	go func() {
		time.Sleep(100 * time.Millisecond)
		common.Log.Debug("Cancelling allocate")
		http.DefaultTransport.(*http.Transport).CancelRequest(req)
	}()

	res, _ := http.DefaultClient.Do(req)
	if res != nil {
		body, _ := ioutil.ReadAll(res.Body)
		require.FailNow(t, "Error: Allocate returned non-nil", string(body))
	}
}
