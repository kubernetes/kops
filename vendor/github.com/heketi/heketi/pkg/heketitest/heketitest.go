//
// Copyright (c) 2016 The heketi Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package heketitest

import (
	"bytes"
	"net/http/httptest"
	"os"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/heketi/heketi/apps/glusterfs"
	"github.com/heketi/heketi/middleware"
	"github.com/heketi/tests"
	"github.com/lpabon/godbc"
)

// Heketi test server configuration
type HeketiMockTestServerConfig struct {
	Auth     bool
	AdminKey string
	UserKey  string
	Logging  bool
}

// Heketi test service metadata
type HeketiMockTestServer struct {
	DbFile string
	Ts     *httptest.Server
	App    *glusterfs.App
}

// Create a simple Heketi mock server
//
// Example:
//     	h := heketitest.NewHeketiMockTestServerDefault()
//		defer h.Close()
//
func NewHeketiMockTestServerDefault() *HeketiMockTestServer {
	return NewHeketiMockTestServer(nil)
}

// Create a Heketi mock server
//
// Example:
//		c := &heketitest.HeketiMockTestServerConfig{
//			Auth:     true,
//			AdminKey: "admin",
//			UserKey:  "user",
// 		    Logging: false,
//		}
//
//		h := heketitest.NewHeketiMockTestServer(c)
//		defer h.Close()
//
func NewHeketiMockTestServer(
	config *HeketiMockTestServerConfig) *HeketiMockTestServer {

	if config == nil {
		config = &HeketiMockTestServerConfig{}
	}

	h := &HeketiMockTestServer{}
	h.DbFile = tests.Tempfile()

	// Set loglevel
	var loglevel string
	if config.Logging {
		loglevel = "debug"
	} else {
		loglevel = "none"
	}

	// Create simple configuration for unit tests
	appConfig := bytes.NewBuffer([]byte(`{
		"glusterfs" : { 
			"executor" : "mock",
			"allocator" : "simple",
			"loglevel" : "` + loglevel + `",
			"db" : "` + h.DbFile + `"
		}
	}`))
	h.App = glusterfs.NewApp(appConfig)
	if h.App == nil {
		return nil
	}

	// Initialize REST service
	h.Ts = h.setupHeketiServer(config)
	if h.Ts == nil {
		return nil
	}

	return h
}

// Get http test service struct
func (h *HeketiMockTestServer) HttpServer() *httptest.Server {
	return h.Ts
}

// Get URL to test server
func (h *HeketiMockTestServer) URL() string {
	return h.Ts.URL
}

// Close database and other services
func (h *HeketiMockTestServer) Close() {
	os.Remove(h.DbFile)
	h.App.Close()
	h.Ts.Close()
}

func (h *HeketiMockTestServer) setupHeketiServer(
	config *HeketiMockTestServerConfig) *httptest.Server {

	godbc.Require(h.App != nil)

	router := mux.NewRouter()
	h.App.SetRoutes(router)
	n := negroni.New()

	// Add authentication
	if config.Auth {
		jwtconfig := &middleware.JwtAuthConfig{}
		jwtconfig.Admin.PrivateKey = config.AdminKey
		jwtconfig.User.PrivateKey = config.UserKey

		// Setup middleware
		n.Use(middleware.NewJwtAuth(jwtconfig))
		n.UseFunc(h.App.Auth)
	}

	// Add App
	n.UseHandler(router)

	// Create server
	return httptest.NewServer(n)
}
