package ui

import (
	"fmt"
	"io"
	"strings"
)

// ConfirmArgs encapsulates the arguments that can he passed to GetConfirm
type ConfirmArgs struct {
	Out     io.Writer // os.Stdout or &bytes.Buffer used to putput the message above the confirmation
	Message string    // what you want to say to the user before confirming
	Default string    // if you hit enter instead of yes or no shoudl it approve or deny
	TestVal string    // if you need to test without the interactive prompt then set the user response here
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
	} else if c.Default != "" {
		if string(c.Default[0]) == "y" {
			return true
		}
		return false
	}
	return GetConfirm(c)
}

// ContainsString returns true if slice contains the element
func ContainsString(slice []string, element string) bool {
	return !(strings.Index(strings.Join(slice, " "), element) == -1)
}
