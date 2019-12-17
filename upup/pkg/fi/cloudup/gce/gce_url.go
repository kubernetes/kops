/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gce

import (
	"fmt"
	"strings"
)

type GoogleCloudURL struct {
	Version string
	Project string
	Type    string
	Name    string
	Global  bool
	Region  string
	Zone    string
}

func (u *GoogleCloudURL) BuildURL() string {
	version := u.Version
	if version == "" {
		version = "v1"
	}
	url := "https://www.googleapis.com/compute/" + version + "/"
	if u.Project != "" {
		url += "projects/" + u.Project + "/"
	}
	if u.Global {
		url += "global/"
	}
	if u.Region != "" {
		url += "regions/" + u.Region + "/"
	}
	if u.Zone != "" {
		url += "zones/" + u.Zone + "/"
	}
	url += u.Type + "/" + u.Name
	return url
}

func ParseGoogleCloudURL(u string) (*GoogleCloudURL, error) {
	tokens := strings.Split(u, "/")
	if len(tokens) < 3 {
		return nil, fmt.Errorf("invalid google cloud URL (token count): %q", u)
	}

	if tokens[0] != "https:" || tokens[1] != "" || tokens[2] != "www.googleapis.com" {
		return nil, fmt.Errorf("invalid google cloud URL (schema / host): %q", u)
	}

	if len(tokens) < 5 || tokens[3] != "compute" {
		return nil, fmt.Errorf("invalid google cloud URL (not compute): %q", u)
	}

	if tokens[4] != "v1" && tokens[4] != "beta" {
		return nil, fmt.Errorf("invalid google cloud URL (not compute/v1 or compute/beta): %q", u)
	}

	parsed := &GoogleCloudURL{
		Version: tokens[4],
	}
	pos := 5
	for {
		if pos >= len(tokens) {
			return nil, fmt.Errorf("invalid google cloud URL (unexpected end): %q", u)
		}
		t := tokens[pos]
		if t == "projects" {
			pos++
			if pos >= len(tokens) {
				return nil, fmt.Errorf("invalid google cloud URL (unexpected projects): %q", u)
			}
			parsed.Project = tokens[pos]
		} else if t == "zones" {
			pos++
			if pos >= len(tokens) {
				return nil, fmt.Errorf("invalid google cloud URL (unexpected zones): %q", u)
			}
			parsed.Zone = tokens[pos]
		} else if t == "regions" && ((pos + 2) < len(tokens)) {
			pos++
			if pos >= len(tokens) {
				return nil, fmt.Errorf("invalid google cloud URL (unexpected regions): %q", u)
			}
			parsed.Region = tokens[pos]
		} else if t == "global" {
			parsed.Global = true
		} else {
			parsed.Type = tokens[pos]
			pos++
			if pos >= len(tokens) {
				return nil, fmt.Errorf("invalid google cloud URL (no name): %q", u)
			}
			parsed.Name = tokens[pos]
			pos++
			if pos != len(tokens) {
				return nil, fmt.Errorf("invalid google cloud URL (content after name): %q", u)
			} else {
				return parsed, nil
			}
		}
		pos++
	}
}
