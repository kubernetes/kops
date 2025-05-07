package ecloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud/schema" 
)

// ------------------------------ API CALLS FUNCTIONS -------------------------

// Login to the API
func (c *Client) Login(reqBody interface{}) (*schema.LoginResponse, error) {
	var res schema.LoginResponse
	err := c.CallAPI("POST", "47777", "/api/v1/authenticate/login", reqBody, &res, false)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Status login
func (c *Client) StatusLogin() (*schema.StatusLoginResponse, error) {
	var res schema.StatusLoginResponse
	err := c.CallAPI("GET", "47777", "/api/v1/authenticate/status", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Logout from the API
func (c *Client) Logout() (*schema.LogoutResponse, error) {
	var res schema.LogoutResponse
	err := c.CallAPI("POST", "47777", "/api/v1/authenticate/logout", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Health check Compute
func (c *Client) HealthCheckCompute() (*schema.HealthCheckComputeResponse, error) {
	var res schema.HealthCheckComputeResponse
	err := c.CallAPI("GET", "17777", "/", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Can allocate a new compute instance
func (c *Client) CanAllocateCompute() (*schema.CanAllocateComputeResponse, error) {
	var res schema.CanAllocateComputeResponse
	err := c.CallAPI("GET", "17777", "/api/v1.0/client/vm/canallocate", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Create a new compute instance
func (c *Client) CreateCompute(reqBody interface{}) (*schema.CreateComputeResponse, error) {
	var res schema.CreateComputeResponse
	err := c.CallAPI("POST", "17777", "/api/v1.0/client/vm/register", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Compute instances status
func (c *Client) ComputeStatus() (*schema.ComputeStatusResponse, error) {
	var res schema.ComputeStatusResponse
	err := c.CallAPI("GET", "17777", "/api/v1.0/client/vm/status", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Compute templates
func (c *Client) ComputeTemplates() (*schema.ComputeTemplatesResponse, error) {
	var res schema.ComputeTemplatesResponse
	err := c.CallAPI("GET", "17777", "/api/v1.0/client/vm/templates", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Compute instance delete
func (c *Client) DeleteCompute(reqBody interface{}) (*schema.DeleteComputeResponse, error) {
	var res schema.DeleteComputeResponse
	err := c.CallAPI("POST", "17777", "/api/v1.0/client/vm/delete", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Health check Storage
func (c *Client) HealthCheckStorage() (*schema.HealthCheckStorageResponse, error) {
	var res schema.HealthCheckStorageResponse
	err := c.CallAPI("GET", "27777", "/", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Can create a new storage volume
func (c *Client) CanCreateStorage() (*schema.CanCreateStorageResponse, error) {
	var res schema.CanCreateStorageResponse
	err := c.CallAPI("GET", "27777", "/api/v1.0/client/volume/cancreate", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Create a new storage volume
func (c *Client) CreateStorage(reqBody interface{}) (*schema.CreateStorageResponse, error) {
	var res schema.CreateStorageResponse
	err := c.CallAPI("POST", "27777", "/api/v1.0/client/volume/create", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Get storages
func (c *Client) GetStorage() (*schema.GetStorageResponse, error) {
	var res schema.GetStorageResponse
	err := c.CallAPI("GET", "27777", "/api/v1.0/client/volume/accessible", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Get storage by ID
func (c *Client) GetStorageByID(reqBody interface{}) (*schema.GetStorageByIDResponse, error) {
	var res schema.GetStorageByIDResponse
	err := c.CallAPI("POST", "27777", "/api/v1.0/client/volume/info", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete a storage volume
func (c *Client) DeleteStorage(reqBody interface{}) (*schema.DeleteStorageResponse, error) {
	var res schema.DeleteStorageResponse
	err := c.CallAPI("POST", "27777", "/api/v1.0/client/volume/delete", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// ------------------------------ UTILS FUNCTIONS -----------------------------

// Base function to perform API calls
func (c *Client) CallAPI(method, port, path string, reqBody, resType interface{}, needAuth bool) error {
	req, err := c.NewRequest(method, port, path, reqBody, needAuth)
	if err != nil {
		return err
	}
	response, err := c.Do(req)
	if err != nil {
		return err
	}
	return c.UnmarshalResponse(response, resType)
}

// NewRequest returns a new HTTP request
func (c *Client) NewRequest(method, port, path string, reqBody interface{}, needAuth bool) (*http.Request, error) {
	var body []byte
	var err error

	if reqBody != nil {
		body, err = json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}
	}

	target := c.endpoint + port + path
	req, err := http.NewRequest(method, target, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
	}
	req.Header.Add("Accept", "application/json")

	if needAuth {
		// TODO: add auth
	}

	c.httpClient.Timeout = c.timeout
	return req, nil
}

// Do sends an HTTP request and returns an HTTP response
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.logger != nil {
		c.logger.LogRequest(req)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if c.logger != nil {
		c.logger.LogResponse(resp)
	}
	return resp, nil
}

// UnmarshalResponse checks the response and unmarshals it into the response type if needed
func (c *Client) UnmarshalResponse(response *http.Response, resType interface{}) error {
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		apiErr := &APIError{
			StatusCode: response.StatusCode,
			Message:    string(body),
		}
		return apiErr
	}

	if len(body) == 0 || resType == nil {
		return nil
	}

	d := json.NewDecoder(bytes.NewReader(body))
	d.UseNumber()
	return d.Decode(resType)
}

// ------------------------------ ERROR HANDLING -----------------------------

type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error: %d - %s", e.StatusCode, e.Message)
}