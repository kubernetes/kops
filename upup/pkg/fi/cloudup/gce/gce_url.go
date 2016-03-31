package gce

import (
	"fmt"
	"strings"
)

type GoogleCloudURL struct {
	Project string
	Type    string
	Name    string
	Global  bool
	Zone    string
}

func (u *GoogleCloudURL) BuildURL() string {
	url := "https://www.googleapis.com/compute/v1/"
	if u.Project != "" {
		url += "projects/" + u.Project + "/"
	}
	if u.Global {
		url += "global/"
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

	if len(tokens) < 5 || tokens[3] != "compute" || tokens[4] != "v1" {
		return nil, fmt.Errorf("invalid google cloud URL (not compute/v1): %q", u)
	}

	parsed := &GoogleCloudURL{}
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
