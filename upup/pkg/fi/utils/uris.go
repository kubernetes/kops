package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func ReadLocation(location string) ([]byte, error) {
	if !strings.Contains(location, "://") {
		// Assume a simple file
		v, err := ioutil.ReadFile(location)
		if err != nil {
			return nil, fmt.Errorf("error reading file %q: %v", location, err)
		}
		return v, nil
	}

	u, err := url.Parse(location)
	if err != nil {
		return nil, fmt.Errorf("error parsing location %q - not a valid URI")
	}

	var httpURL string
	httpHeaders := make(map[string]string)

	switch u.Scheme {
	case "metadata":
		switch u.Host {
		case "gce":
			httpURL = "http://169.254.169.254/computeMetadata/v1/instance/attributes/" + u.Path
			httpHeaders["Metadata-Flavor"] = "Google"

		case "aws":
			httpURL = "http://169.254.169.254/latest/" + u.Path

		default:
			return nil, fmt.Errorf("unknown metadata type: %q in %q", u.Host, location)
		}

	case "http", "https":
		httpURL = location

	default:
		return nil, fmt.Errorf("unrecognized scheme for location %q")
	}

	req, err := http.NewRequest("GET", httpURL, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range httpHeaders {
		req.Header.Add(k, v)
	}
	response, err := http.DefaultClient.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching %q: %v", httpURL, err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response for %q: %v", httpURL, err)
	}
	return body, nil
}
