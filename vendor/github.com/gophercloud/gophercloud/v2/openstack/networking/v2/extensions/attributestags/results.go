package attributestags

import (
	"net/http"

	"github.com/gophercloud/gophercloud/v2"
)

type tagResult struct {
	gophercloud.Result
}

// Extract interprets tagResult to return the list of tags
func (r tagResult) Extract() ([]string, error) {
	var s struct {
		Tags []string `json:"tags"`
	}
	err := r.ExtractInto(&s)
	return s.Tags, err
}

// ReplaceAllResult represents the result of a replace operation.
// Call its Extract method to interpret it as a slice of strings.
type ReplaceAllResult struct {
	tagResult
}

type ListResult struct {
	tagResult
}

// DeleteResult is the result from a Delete/DeleteAll operation.
// Call its ExtractErr method to determine if the call succeeded or failed.
type DeleteResult struct {
	gophercloud.ErrResult
}

// AddResult is the result from an Add operation.
// Call its ExtractErr method to determine if the call succeeded or failed.
type AddResult struct {
	gophercloud.ErrResult
}

// ConfirmResult is the result from an Confirm operation.
type ConfirmResult struct {
	gophercloud.Result
}

func (r ConfirmResult) Extract() (bool, error) {
	exists := r.Err == nil

	if gophercloud.ResponseCodeIs(r.Err, http.StatusNotFound) {
		r.Err = nil
	}

	return exists, r.Err
}
