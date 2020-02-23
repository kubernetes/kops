package useragent

import (
	"fmt"
	"strings"
)

// UserAgent represents a User-Agent header.
type UserAgent struct {
	// Product identifier; its name or development codename.
	Product string `json:"product"`
	// Version number of the product.
	Version string `json:"version"`
	// Zero or more comments containing more details.
	Comment []string `json:"comment"`
}

// UserAgents represents one or more UserAgents.
type UserAgents []UserAgent

// New returns a UserAgent.
func New(product, version string, comment ...string) UserAgent {
	return UserAgent{
		Product: product,
		Version: version,
		Comment: comment,
	}
}

// String returns the string representation of UserAgent.
func (ua UserAgent) String() string {
	s := fmt.Sprintf("%s/%s", ua.Product, ua.Version)

	if len(ua.Comment) > 0 {
		s += fmt.Sprintf(" (%s)", strings.Join(ua.Comment, "; "))
	}

	return s
}

// String concatenates all the user-defined UserAgents.
func (uas UserAgents) String() string {
	ss := make([]string, len(uas))

	for i, ua := range uas {
		ss[i] = ua.String()
	}

	return strings.Join(ss, " ")
}
