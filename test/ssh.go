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
	"os"
	"io/ioutil"
	"path/filepath"
	"fmt"
	"strings"
)

var (
	// Hermione requires a PUBLIC SSH key to drop off on the cluster for access..
	// The private half of this key is carefully hidden in a place most users will never know about
	DefaultSSHPublicKeyBytes = []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCnhkq2fUyDMImFAfiRItxuVNbcLc6CTj0Fipb/DWlNBhVSL34fJeUw4oFqPDuL0N2cUTsxH22drie/AOJCVzZ4f7OEECFAh6RD2I11GxqrlDppr4KjWBoV//XyGbPZa4NXg+Bpg0+UVefqjAQNiAIRmuJ8vBsl952Vkf46RIKNAznoNUI7vnYkfvU1eCCaFcWiysxfg8BALiKS4Trv00TmHO15PPuiHeYzeu4wAB+0Mtuc6uh4WmqIWOEe/jmIhdHXYniHm3AEtMjL1GVMly/QEKjANUFZCZsBFbm3S6T+o820reOXSSG1wW2XBfOSTNyGlBZJFnutQntBwvHpw6A9 kris@Goddess-Of-Production.local")

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
