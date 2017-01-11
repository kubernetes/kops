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

package ui

import (
	"fmt"
	"io"
	"strings"
)

// ConfirmArgs encapsulates the arguments that can he passed to GetConfirm
type ConfirmArgs struct {
	Out        io.Writer // os.Stdout or &bytes.Buffer used to putput the message above the confirmation
	Message    string    // what you want to say to the user before confirming
	Default    string    // if you hit enter instead of yes or no shoudl it approve or deny
	TestVal    string    // if you need to test without the interactive prompt then set the user response here
	Retries    int       // how many tines to ask for a valid confirmation before giving up
	RetryCount int       // how many attempts have been made
}

// GetConfirm prompts a user for a yes or no answer.
// In order to test this function som extra parameters are reqired:
//
// out: an io.Writer that allows you to direct prints to stdout or another location
// message: the string that will be printed just before prompting for a yes or no.
// answer: "", "yes", or "no" - this allows for easier testing
func GetConfirm(c *ConfirmArgs) bool {
	if c.Default != "" {
		c.Default = strings.ToLower(c.Default)
	}
	answerTemplate := "(%s/%s)"
	switch c.Default {
	case "yes", "y":
		c.Message = c.Message + fmt.Sprintf(answerTemplate, "Y", "n")
	case "no", "n":
		c.Message = c.Message + fmt.Sprintf(answerTemplate, "y", "N")
	default:
		c.Message = c.Message + fmt.Sprintf(answerTemplate, "y", "n")
	}
	fmt.Fprintln(c.Out, c.Message)

	// these are the acceptable answers
	okayResponses := []string{"y", "yes"}
	nokayResponses := []string{"n", "no"}
	response := c.TestVal

	// only prompt user if you predefined answer was passed in
	if response == "" {
		_, err := fmt.Scanln(&response)
		if err != nil {
			return false
		}
	}

	responseLower := strings.ToLower(response)
	// make sure the response is valid
	if ContainsString(okayResponses, responseLower) {
		return true
	} else if ContainsString(nokayResponses, responseLower) {
		return false
	} else if c.Default != "" && response == "" {
		if string(c.Default[0]) == "y" {
			return true
		}
		return false
	}

	fmt.Printf("invalid response: %s\n\n", response)

	// if c.RetryCount exceeds the requested number of retries then five up
	if c.RetryCount >= c.Retries {
		return false
	}

	c.RetryCount++
	return GetConfirm(c)
}

// ContainsString returns true if slice contains the element
func ContainsString(slice []string, element string) bool {
	for _, arg := range slice {
		if arg == element {
			return true
		}
	}
	return false
}
