package ecloud

import (
	"net/http"
	"encoding/json"
	"bytes"
	"io"
)


// ------------------------------ API CALLS FUNCTIONS -------------------------

// Login to the API
func (c *Client) Login(reqBody interface{}, resType interface{}) error {
	return c.CallAPI("POST", "47777", "/api/v1/authenticate/login", reqBody, resType, false)
}

// Status login
func (c *Client) StatusLogin(resType interface{}) error {
	return c.CallAPI("GET", "47777", "/api/v1/authenticate/status", nil, resType, true)
}

// Logout from the API
func (c *Client) Logout(resType interface{}) error {
	return c.CallAPI("POST", "47777", "/api/v1/authenticate/logout", nil, resType, true)
}

// Health check Compute
func (c *Client) HealthCheckCompute(resType interface{}) error {
	return c.CallAPI("GET", "17777", "/", nil, resType, true)
}

// Can allocate a new compute instance
func (c *Client) CanAllocateCompute(resType interface{}) error {
	return c.CallAPI("GET", "17777", "/api/v1.0/client/vm/canallocate", nil, resType, true)
}

// Create a new compute instance
func (c *Client) CreateCompute(reqBody interface{}, resType interface{}) error {
	return c.CallAPI("POST", "17777", "/api/v1.0/client/vm/register", reqBody, resType, true)
}

// Compute instance status
func (c *Client) ComputeStatus(resType interface{}) error {
	return c.CallAPI("GET", "17777", "/api/v1.0/client/vm/status", nil, resType, true)
}

// Compute templates
func (c *Client) ComputeTemplates(resType interface{}) error {
	return c.CallAPI("GET", "17777", "/api/v1.0/client/vm/templates", nil, resType, true)
}

// Compute instance delete
func (c *Client) DeleteCompute(reqBody interface{}, resType interface{}) error {
	return c.CallAPI("POST", "17777", "/api/v1.0/client/vm/delete", reqBody, resType, true)
}

// Health check Storage
func (c *Client) HealthCheckStorage(resType interface{}) error {
	return c.CallAPI("GET", "27777", "/", nil, resType, true)
}

// Can allocate a new storage volume
func (c *Client) CanAllocateStorage(resType interface{}) error {
	return c.CallAPI("GET", "27777", "/api/v1.0/client/volume/cancreate", nil, resType, true)
}

// Create a new storage volume
func (c *Client) CreateStorage(reqBody interface{}, resType interface{}) error {
	return c.CallAPI("POST", "27777", "/api/v1.0/client/volume/create", reqBody, resType, true)
}

// Get storages
func (c *Client) GetStorage(resType interface{}) error {
	return c.CallAPI("GET", "27777", "/api/v1.0/client/volume/accessible", nil, resType, true)
}

// Get storage by ID
func (c *Client) GetStorageByID(reqBody interface{}, resType interface{}) error {
	return c.CallAPI("POST", "27777", "/api/v1.0/client/volume/info", reqBody, resType, true)
}

// Delete a storage volume
func (c *Client) DeleteStorage(reqBody interface{}, resType interface{}) error {
	return c.CallAPI("POST", "27777", "/api/v1.0/client/volume/delete", reqBody, resType, true)
}

// TODO: Network SDN endpoints...


// ------------------------------ UTILS FUNCTIONS -----------------------------

// Base function to perform API calls
// Args:
// - method: HTTP method to use
// - path: API path to call
// - reqBody: request body to send
// - resType: response type to unmarshal
// - needAuth: if the call needs authentication
// Returns:
// - error: if any
func (c *Client) CallAPI(method, port string, path string, reqBody, resType interface{}, needAuth bool) error {
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
func (c *Client) NewRequest(method, port string, path string, reqBody interface{}, needAuth bool) (*http.Request, error) {
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

	// Inject headers
	// TODO: insert real headers
	if body != nil {
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
	}
	req.Header.Add("Accept", "application/json")

	// Inject signature. Some methods do not need authentication, especially /time,
	// /auth and some /order methods are actually broken if authenticated.
	if needAuth {
		// TODO: insert auth process
	}

	// Send the request with requested timeout
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


// UnmarshalResponse checks the response and unmarshals it into the response
// type if needed Helper function, called from CallAPI
func (c *Client) UnmarshalResponse(response *http.Response, resType interface{}) error {
	// Read all the response body
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// < 200 && >= 300 then generate API error
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		// TODO: decide how to handle errors
		// apiError := &APIError{Code: response.StatusCode}
		// if err = json.Unmarshal(body, apiError); err != nil {
		// 	apiError.Message = string(body)
		// }
		// apiError.QueryID = response.Header.Get("X-Ovh-QueryID")

		// return apiError
	}

	// Nothing to unmarshal
	if len(body) == 0 || resType == nil {
		return nil
	}

	d := json.NewDecoder(bytes.NewReader(body))
	d.UseNumber()
	return d.Decode(&resType)
}