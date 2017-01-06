package ui

import (
	"fmt"
	"io"
	"strings"
)

// GetConfirm prompts a user for a yes or no answer.
// In order to test this function som extra parameters are reqired:
//
// out: an io.Writer that allows you to direct prints to stdout or another location
// message: the string that will be printed just before prompting for a yes or no.
// answer: "", "yes", or "no" - this allows for easier testing
func GetConfirm(out io.Writer, message string, answer string) bool {
	fmt.Fprintln(out, message)

	// these are the acceptable answers
	okayResponses := []string{"y", "yes"}
	nokayResponses := []string{"n", "no"}
	response := answer

	// only prompt user if you predefined answer was passed in
	if answer == "" {
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
	} else {
		return GetConfirm(out, "Please type yes or no and then press enter:", answer)
	}
}

// ContainsString returns true if slice contains the element
func ContainsString(slice []string, element string) bool {
	return !(strings.Index(strings.Join(slice, " "), element) == -1)
}
