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
	"bufio"
	"fmt"
	"io"
	"k8s.io/apimachinery/pkg/util/sets"
	"os"
	"strings"
)

// ConfirmArgs encapsulates the arguments that can he passed to GetConfirm
type ConfirmArgs struct {
	Out        io.Writer // os.Stdout or &bytes.Buffer used to output the message above the confirmation
	Message    string    // what you want to say to the user before confirming
	Default    string    // if you hit enter instead of yes or no should it approve or deny
	TestVal    string    // if you need to test without the interactive prompt then set the user response here
	Retries    int       // how many tines to ask for a valid confirmation before giving up
	retryCount int       // how many attempts have been made
}

// GetConfirm prompts a user for a yes or no answer.
// In order to test this function some extra parameters are required:
//
// out: an io.Writer that allows you to direct prints to stdout or another location
// message: the string that will be printed just before prompting for a yes or no.
// answer: "", "yes", or "no" - this allows for easier testing
func GetConfirm(c *ConfirmArgs) (bool, error) {
	if c.Default != "" {
		c.Default = strings.ToLower(c.Default)
	}

	for {
		answerTemplate := " (%s/%s)"
		message := c.Message
		switch c.Default {
		case "yes", "y":
			message = c.Message + fmt.Sprintf(answerTemplate, "Y", "n")
		case "no", "n":
			message = c.Message + fmt.Sprintf(answerTemplate, "y", "N")
		default:
			message = c.Message + fmt.Sprintf(answerTemplate, "y", "n")
		}
		fmt.Fprintln(c.Out, message)

		// these are the acceptable answers
		okayResponses := sets.NewString("y", "yes")
		nokayResponses := sets.NewString("n", "no")
		response := c.TestVal

		// only prompt user if no predefined answer was passed in
		if response == "" {
			var err error

			reader := bufio.NewReader(os.Stdin)
			response, err = reader.ReadString('\n')
			if err != nil {
				return false, fmt.Errorf("error reading from input: %v", err)
			}

			response = strings.TrimSpace(response)
		}

		responseLower := strings.ToLower(response)
		// make sure the response is valid
		if okayResponses.Has(responseLower) {
			return true, nil
		} else if nokayResponses.Has(responseLower) {
			return false, nil
		} else if c.Default != "" && response == "" {
			if string(c.Default[0]) == "y" {
				return true, nil
			}
			return false, nil
		}

		fmt.Printf("invalid response: %s\n\n", response)

		// if c.RetryCount exceeds the requested number of retries then give up
		if c.retryCount >= c.Retries {
			return false, nil
		}

		c.retryCount++
	}
}
