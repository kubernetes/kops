// Package validation provides format validation functions.
package validation

import (
	"net/url"
	"regexp"
)

var (
	isUUIDRegexp  = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")
	isRegionRegex = regexp.MustCompile("^[a-z]{2}-[a-z]{3}$")
	isZoneRegex   = regexp.MustCompile("^[a-z]{2}-[a-z]{3}-[1-9]$")
	isAccessKey   = regexp.MustCompile("^SCW[A-Z0-9]{17}$")
	isEmailRegexp = regexp.MustCompile("^.+@.+$")
)

// IsUUID returns true if the given string has a valid UUID format.
func IsUUID(s string) bool {
	return isUUIDRegexp.MatchString(s)
}

// IsAccessKey returns true if the given string has a valid Scaleway access key format.
func IsAccessKey(s string) bool {
	return isAccessKey.MatchString(s)
}

// IsSecretKey returns true if the given string has a valid Scaleway secret key format.
func IsSecretKey(s string) bool {
	return IsUUID(s)
}

// IsOrganizationID returns true if the given string has a valid Scaleway organization ID format.
func IsOrganizationID(s string) bool {
	return IsUUID(s)
}

// IsProjectID returns true if the given string has a valid Scaleway project ID format.
func IsProjectID(s string) bool {
	return IsUUID(s)
}

// IsRegion returns true if the given string has a valid region format.
func IsRegion(s string) bool {
	return isRegionRegex.MatchString(s)
}

// IsZone returns true if the given string has a valid zone format.
func IsZone(s string) bool {
	return isZoneRegex.MatchString(s)
}

// IsURL returns true if the given string has a valid URL format.
func IsURL(s string) bool {
	_, err := url.Parse(s)
	return err == nil
}

// IsEmail returns true if the given string has an email format.
func IsEmail(v string) bool {
	return isEmailRegexp.MatchString(v)
}
