/*
Copyright 2016 The Kubernetes Authors.

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

package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	// kops requires a PUBLIC SSH key to drop off on the cluster for access..
	// The private half of this key is carefully hidden in a place most users will never know about
	DefaultSSHPublicKeyBytes = []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDVAkDsCA4Hdwpw5iT+YyyuCjjHgIUmcTRLN+nYbLeBlgq2hibXaNFNBwLXMk3DNN4Rr9keItHK17Wikij2hV5XoZwO3Dob5QROnYEzFh1JKPah93HOeoXbQ3mBuy3yz4iw7xMPxG5GGsXJFZYEQGEze5NHyg0Gz3dHt3djn8WXvQvR9F6tUOqTF0y4FVo3gMjVJQsvQaL5iJ8Hdxw5djK1SNa9FN2kJ6nwRlkeTIWqRVOMQ60aiIouh6g1wO0lrfcj68JgNKOM7KXsb7E02twCWOyvupH7tOqsWcO3oF5R8wCG11eoBza1UIKpvnDSeijubMMlfZyiaKLb9eg1SU4L clove@clove-mbp.local")
)

func EnsurePublicKey(keyLocation string) error {
	if !strings.Contains(keyLocation, ".pub") {
		return fmt.Errorf("Must use a public key in format /path/to/key.pub")
	}
	if _, err := os.Stat(keyLocation); os.IsNotExist(err) {
		// We need to create the key
		err = os.MkdirAll(filepath.Dir(keyLocation), 0700)
		if err != nil {
			return fmt.Errorf("Unable to create directory %s", filepath.Dir(keyLocation))
		}
		err = ioutil.WriteFile(keyLocation, DefaultSSHPublicKeyBytes, 0600)
		if err != nil {
			return fmt.Errorf("Unable to write file %s", keyLocation)
		}
	}
	return nil
}
