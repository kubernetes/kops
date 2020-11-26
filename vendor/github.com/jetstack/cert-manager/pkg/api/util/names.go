/*
Copyright 2019 The Jetstack cert-manager contributors.

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

package util

import (
	"encoding/json"
	"fmt"
	"hash/fnv"

	"regexp"
)

func ComputeName(prefix string, obj interface{}) (string, error) {
	objectBytes, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	hashF := fnv.New32()
	_, err = hashF.Write(objectBytes)
	if err != nil {
		return "", err
	}

	// we're shortening to stay under 64 as we use this in services
	// and pods down the road for ACME resources.
	prefix = DNSSafeShortenTo52Characters(prefix)

	return fmt.Sprintf("%s-%d", prefix, hashF.Sum32()), nil
}

func DNSSafeShortenTo52Characters(in string) string {
	if len(in) >= 52 {
		// shorten the cert name to 52 chars to ensure the total length of the name
		// also shorten the 52 char string to the last non-symbol character
		// is less than or equal to 64 characters
		validCharIndexes := regexp.MustCompile(`[a-zA-Z\d]`).FindAllStringIndex(fmt.Sprintf("%.52s", in), -1)
		in = in[:validCharIndexes[len(validCharIndexes)-1][1]]
	}

	return in
}
