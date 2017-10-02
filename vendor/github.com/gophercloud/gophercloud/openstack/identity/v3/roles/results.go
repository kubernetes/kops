package roles

import "github.com/gophercloud/gophercloud/pagination"

// RoleAssignment is the result of a role assignments query.
type RoleAssignment struct {
	Role  Role  `json:"role,omitempty"`
	Scope Scope `json:"scope,omitempty"`
	User  User  `json:"user,omitempty"`
	Group Group `json:"group,omitempty"`
}

// Role represents a Role in an assignment.
type Role struct {
	ID string `json:"id,omitempty"`
}

// Scope represents a scope in a Role assignment.
type Scope struct {
	Domain  Domain  `json:"domain,omitempty"`
	Project Project `json:"project,omitempty"`
}

// Domain represents a domain in a role assignment scope.
type Domain struct {
	ID string `json:"id,omitempty"`
}

// Project represents a project in a role assignment scope.
type Project struct {
	ID string `json:"id,omitempty"`
}

// User represents a user in a role assignment scope.
type User struct {
	ID string `json:"id,omitempty"`
}

// Group represents a group in a role assignment scope.
type Group struct {
	ID string `json:"id,omitempty"`
}

// RoleAssignmentPage is a single page of RoleAssignments results.
type RoleAssignmentPage struct {
	pagination.LinkedPageBase
}

// IsEmpty returns true if the RoleAssignmentPage contains no results.
func (r RoleAssignmentPage) IsEmpty() (bool, error) {
	roleAssignments, err := ExtractRoleAssignments(r)
	return len(roleAssignments) == 0, err
}

// NextPageURL uses the response's embedded link reference to navigate to
// the next page of results.
func (r RoleAssignmentPage) NextPageURL() (string, error) {
	var s struct {
		Links struct {
			Next string `json:"next"`
		} `json:"links"`
	}
	err := r.ExtractInto(&s)
	return s.Links.Next, err
}

// ExtractRoleAssignments extracts a slice of RoleAssignments from a Collection
// acquired from List.
func ExtractRoleAssignments(r pagination.Page) ([]RoleAssignment, error) {
	var s struct {
		RoleAssignments []RoleAssignment `json:"role_assignments"`
	}
	err := (r.(RoleAssignmentPage)).ExtractInto(&s)
	return s.RoleAssignments, err
}
