// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// How many seconds the program should wait before trying to connect to the dashboard again
const RetryTimeout = 5

type grafanaConfig struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Access    string `json:"access"`
	IsDefault bool   `json:"isDefault"`
	URL       string `json:"url"`
	Password  string `json:"password"`
	User      string `json:"user"`
	Database  string `json:"database"`
}

func main() {

	envParams := map[string]string{
		"grafana_user":              "admin",
		"grafana_passwd":            "admin",
		"grafana_port":              "3000",
		"influxdb_host":             "monitoring-influxdb",
		"influxdb_port":             "8086",
		"influxdb_database":         "k8s",
		"influxdb_user":             "root",
		"influxdb_password":         "root",
		"influxdb_service_url":      "",
		"dashboard_location":        "/dashboards",
		"gf_auth_anonymous_enabled": "true",
		"gf_server_protocol":        "http",
		"backend_access_mode":       "proxy",
	}

	for k := range envParams {
		if v := os.Getenv(strings.ToUpper(k)); v != "" {
			envParams[k] = v
		}
	}

	if envParams["influxdb_service_url"] == "" {
		envParams["influxdb_service_url"] = fmt.Sprintf("http://%s:%s", envParams["influxdb_host"], envParams["influxdb_port"])
	}

	cfg := grafanaConfig{
		Name:      "influxdb-datasource",
		Type:      "influxdb",
		Access:    envParams["backend_access_mode"],
		IsDefault: true,
		URL:       envParams["influxdb_service_url"],
		User:      envParams["influxdb_user"],
		Password:  envParams["influxdb_password"],
		Database:  envParams["influxdb_database"],
	}

	grafanaURL := fmt.Sprintf("%s://%s:%s@localhost:%s", envParams["gf_server_protocol"], envParams["grafana_user"], envParams["grafana_passwd"], envParams["grafana_port"])

	for {
		res, err := http.Get(grafanaURL + "/api/org")
		if err != nil {
			fmt.Printf("Can't access the Grafana dashboard. Error: %v. Retrying after %d seconds...\n", err, RetryTimeout)
			time.Sleep(RetryTimeout * time.Second)
			continue
		}

		_, err = ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			fmt.Printf("Can't access the Grafana dashboard. Error: %v. Retrying after %d seconds...\n", err, RetryTimeout)
			time.Sleep(RetryTimeout * time.Second)
			continue
		}

		fmt.Println("Connected to the Grafana dashboard.")
		break
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(cfg)

	for {
		_, err := http.Post(grafanaURL+"/api/datasources", "application/json; charset=utf-8", b)
		if err != nil {
			fmt.Printf("Failed to configure the Grafana dashboard. Error: %v. Retrying after %d seconds...\n", err, RetryTimeout)
			time.Sleep(RetryTimeout * time.Second)
			continue
		}

		fmt.Println("The datasource for the Grafana dashboard is now set.")
		break
	}

	dashboardDir := envParams["dashboard_location"]
	files, err := ioutil.ReadDir(dashboardDir)
	if err != nil {
		fmt.Printf("Failed to read the the directory the json files should be in. Exiting... Error: %v\n", err)
		os.Exit(1)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(dashboardDir, file.Name())
		jsonbytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Failed to read the json file: %s. Proceeding with the next one. Error: %v\n", filePath, err)
			continue
		}

		_, err = http.Post(grafanaURL+"/api/dashboards/db", "application/json; charset=utf-8", bytes.NewReader(jsonbytes))
		if err != nil {
			fmt.Printf("Failed to post the json file: %s. Proceeding with the next one. Error: %v\n", filePath, err)
			continue
		}
	}
}
