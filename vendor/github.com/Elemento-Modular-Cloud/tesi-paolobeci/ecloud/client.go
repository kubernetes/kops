package ecloud

// TODO: fallo simile ad hetzner

import (
	"fmt"
	"net/http"
	"time"
)

type Response struct {
	*http.Response
	// Meta Meta TODO

	// body holds a copy of the http.Response body that must be used within the handler
	// chain. The http.Response.Body is reserved for external users.
	body []byte
}

// Meson endpoints of supportd Cloud Providers
const (
	ArubaCloudEU   = "http://192.168.1.103" // = "https://eu.arubacloud.public.elemento.cloud/api/v1.0"
	OvhEU          = "http://192.168.1.103" // = "https://eu.ovh.public.elemento.cloud/api/v1.0"
	GigasEU        = "http://192.168.1.103" // = "https://eu.gigas.public.elemento.cloud/api/v1.0"
	IonosEU        = "http://192.168.1.103" // = "https://eu.ionos.public.elemento.cloud/api/v1.0"
)

var Endpoints = map[string]string{
	"arubacloud-eu":   ArubaCloudEU,
	"ovh-eu":          OvhEU,
	"gigas-eu":        GigasEU,
	"ionos-eu":        IonosEU,
}


// Local endpoints for deamons connection
const (
	AuthenticateRoute = "http://localhost:47777/api/v1/authenticate/"
	ComputeRoute	  = "http://localhost:17777/api/v1.0/client/vm/"
	StorageRoute	  = "http://localhost:27777/api/v1.0/client/volume/"
	NetworkRoute	  = "http://localhost:37777/"
)

var Routes = map[string]string{
	"authenticate": AuthenticateRoute,
	"compute":      ComputeRoute,
	"storage":      StorageRoute,
	"network":      NetworkRoute,
}

// Client represents a client to call the Elemento Cloud API
type Client struct {
	endpoint                string
	retryMaxRetries         int
	timeout 			    time.Duration
	httpClient              *http.Client
	applicationName         string
	applicationVersion      string
	userAgent               string
	logger 					Logger

	// TODO

	// Firewall         FirewallClient
	// Location         LocationClient
	// Network          NetworkClient
	// Pricing          PricingClient
	// SSHKey           SSHKeyClient
	// Logger 			 Logger
}

func NewClient(applicationName string, applicationVersion string) (*Client, error) {
	if applicationName == "" {
		return nil, fmt.Errorf("application name cannot be empty")
	}
	if applicationVersion == "" {
		return nil, fmt.Errorf("application version cannot be empty")
	}

	client := &Client{
		endpoint:           "http://192.168.1.19",
		retryMaxRetries:    3,
		timeout:            30 * time.Second,
		httpClient:         &http.Client{},
		applicationName:    applicationName,
		applicationVersion: applicationVersion,
		userAgent:          fmt.Sprintf("%s/%s", applicationName, applicationVersion),
	}

	// TODO: research real data needed for the client

	return client, nil
}

// New Client creation from env variables
func NewClientFromEnv(path string) (*Client, error) {
	var client Client

	// Get and check the configuration
	if err := client.loadConfig(path); err != nil {
		return nil, err
	}
	return &client, nil
}