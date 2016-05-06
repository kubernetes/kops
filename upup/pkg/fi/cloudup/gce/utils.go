package gce

import (
	"google.golang.org/api/googleapi"
)

func IsNotFound(err error) bool {
	apiErr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}

	// We could also check for Errors[].Resource == "notFound"
	//glog.Info("apiErr: %v", apiErr)

	return apiErr.Code == 404
}

func IsNotReady(err error) bool {
	apiErr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}
	for _, e := range apiErr.Errors {
		if e.Reason == "resourceNotReady" {
			return true
		}
	}
	return false
}
