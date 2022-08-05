package hcloud

import (
	"fmt"
	"regexp"
)

var keyRegexp = regexp.MustCompile(
	`^([a-z0-9A-Z]((?:[\-_.]|[a-z0-9A-Z]){0,253}[a-z0-9A-Z])?/)?[a-z0-9A-Z]((?:[\-_.]|[a-z0-9A-Z]|){0,62}[a-z0-9A-Z])?$`)
var valueRegexp = regexp.MustCompile(`^(([a-z0-9A-Z](?:[\-_.]|[a-z0-9A-Z]){0,62})?[a-z0-9A-Z]$|$)`)

func ValidateResourceLabels(labels map[string]interface{}) (bool, error) {
	for k, v := range labels {
		if match := keyRegexp.MatchString(k); !match {
			return false, fmt.Errorf("label key '%s' is not correctly formatted", k)
		}

		if match := valueRegexp.MatchString(v.(string)); !match {
			return false, fmt.Errorf("label value '%s' (key: %s) is not correctly formatted", v, k)
		}
	}
	return true, nil
}
