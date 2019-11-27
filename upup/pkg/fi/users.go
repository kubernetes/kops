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

package fi

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"k8s.io/klog"
)

// This file parses /etc/passwd and /etc/group to get information about users & groups
// Go has built-in user functionality, and group functionality is merged but not yet released
// TODO: Replace this file with e.g. user.LookupGroup once 42f07ff2679d38a03522db3ccd488f4cc230c8c2 lands in go 1.7

type User struct {
	Name    string
	Uid     int
	Gid     int
	Comment string
	Home    string
	Shell   string
}

func parseUsers() (map[string]*User, error) {
	users := make(map[string]*User)

	path := "/etc/passwd"
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading user file %q", path)
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		tokens := strings.Split(line, ":")

		if len(tokens) < 7 {
			klog.Warningf("Ignoring malformed /etc/passwd line (too few tokens): %q\n", line)
			continue
		}

		uid, err := strconv.Atoi(tokens[2])
		if err != nil {
			klog.Warningf("Ignoring malformed /etc/passwd line (bad uid): %q", line)
			continue
		}
		gid, err := strconv.Atoi(tokens[3])
		if err != nil {
			klog.Warningf("Ignoring malformed /etc/passwd line (bad gid): %q", line)
			continue
		}

		u := &User{
			Name: tokens[0],
			// Password
			Uid:     uid,
			Gid:     gid,
			Comment: tokens[4],
			Home:    tokens[5],
			Shell:   tokens[6],
		}
		users[u.Name] = u
	}
	return users, nil
}

func LookupUser(name string) (*User, error) {
	users, err := parseUsers()
	if err != nil {
		return nil, fmt.Errorf("error reading users: %v", err)
	}
	return users[name], nil
}

func LookupUserById(uid int) (*User, error) {
	users, err := parseUsers()
	if err != nil {
		return nil, fmt.Errorf("error reading users: %v", err)
	}
	for _, v := range users {
		if v.Uid == uid {
			return v, nil
		}
	}
	return nil, nil
}

type Group struct {
	Name string
	Gid  int
	//Members []string
}

func parseGroups() (map[string]*Group, error) {
	groups := make(map[string]*Group)

	path := "/etc/group"
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading group file %q", path)
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		tokens := strings.Split(line, ":")

		if len(tokens) < 4 {
			klog.Warningf("Ignoring malformed /etc/group line (too few tokens): %q", line)
			continue
		}

		gid, err := strconv.Atoi(tokens[2])
		if err != nil {
			klog.Warningf("Ignoring malformed /etc/group line (bad gid): %q", line)
			continue
		}

		g := &Group{
			Name: tokens[0],
			// Password: tokens[1]
			Gid: gid,
			// Members: strings.Split(tokens[3], ",")
		}
		groups[g.Name] = g
	}
	return groups, nil
}

func LookupGroup(name string) (*Group, error) {
	groups, err := parseGroups()
	if err != nil {
		return nil, fmt.Errorf("error reading groups: %v", err)
	}
	return groups[name], nil
}

func LookupGroupById(gid int) (*Group, error) {
	users, err := parseGroups()
	if err != nil {
		return nil, fmt.Errorf("error reading groups: %v", err)
	}
	for _, v := range users {
		if v.Gid == gid {
			return v, nil
		}
	}
	return nil, nil
}
