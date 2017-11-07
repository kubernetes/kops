package services

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

type commonResult struct {
	gophercloud.Result
}

// Extract interprets a GetResult, CreateResult or UpdateResult as a concrete
// Service. An error is returned if the original call or the extraction failed.
func (r commonResult) Extract() (*Service, error) {
	var s struct {
		Service *Service `json:"service"`
	}
	err := r.ExtractInto(&s)
	return s.Service, err
}

// CreateResult is the response from a Create request. Call its Extract method
// to interpret it as a Service.
type CreateResult struct {
	commonResult
}

// GetResult is the response from a Get request. Call its Extract method
// to interpret it as a Service.
type GetResult struct {
	commonResult
}

// UpdateResult is the response from an Update request. Call its Extract method
// to interpret it as a Service.
type UpdateResult struct {
	commonResult
}

// DeleteResult is the response from a Delete request. Call its ExtractErr
// method to interpret it as a Service.
type DeleteResult struct {
	gophercloud.ErrResult
}

// Service represents an OpenStack Service.
type Service struct {
	// Description is a description of the service.
	Description string `json:"description"`

	// ID is the unique ID of the service.
	ID string `json:"id"`

	// Name is the name of the service.
	Name string `json:"name"`

	// Type is the type of the service.
	Type string `json:"type"`
}

// ServicePage is a single page of Service results.
type ServicePage struct {
	pagination.LinkedPageBase
}

// IsEmpty returns true if the ServicePage contains no results.
func (p ServicePage) IsEmpty() (bool, error) {
	services, err := ExtractServices(p)
	return len(services) == 0, err
}

// ExtractServices extracts a slice of Services from a Collection acquired
// from List.
func ExtractServices(r pagination.Page) ([]Service, error) {
	var s struct {
		Services []Service `json:"services"`
	}
	err := (r.(ServicePage)).ExtractInto(&s)
	return s.Services, err
}
