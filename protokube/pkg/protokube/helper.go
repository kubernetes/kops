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

package protokube

import (
	"fmt"
	"os"
)

// touchFile does what is says on the tin, it touches a file
func touchFile(p string) error {
	_, err := os.Lstat(p)
	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("error getting state of file %q: %v", p, err)
	}

	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("error touching file %q: %v", p, err)
	}

	if err = f.Close(); err != nil {
		return fmt.Errorf("error closing touched file %q: %v", p, err)
	}

	return nil
}
