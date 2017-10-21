package roles

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

// ListAssignmentsOptsBuilder allows extensions to add additional parameters to
// the ListAssignments request.
type ListAssignmentsOptsBuilder interface {
	ToRolesListAssignmentsQuery() (string, error)
}

// ListAssignmentsOpts allows you to query the ListAssignments method.
// Specify one of or a combination of GroupId, RoleId, ScopeDomainId,
// ScopeProjectId, and/or UserId to search for roles assigned to corresponding
// entities.
type ListAssignmentsOpts struct {
	// GroupID is the group ID to query.
	GroupID string `q:"group.id"`

	// RoleID is the specific role to query assignments to.
	RoleID string `q:"role.id"`

	// ScopeDomainID filters the results by the given domain ID.
	ScopeDomainID string `q:"scope.domain.id"`

	// ScopeProjectID filters the results by the given Project ID.
	ScopeProjectID string `q:"scope.project.id"`

	// UserID filterst he results by the given User ID.
	UserID string `q:"user.id"`

	// Effective lists effective assignments at the user, project, and domain
	// level, allowing for the effects of group membership.
	Effective *bool `q:"effective"`
}

// ToRolesListAssignmentsQuery formats a ListAssignmentsOpts into a query string.
func (opts ListAssignmentsOpts) ToRolesListAssignmentsQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

// ListAssignments enumerates the roles assigned to a specified resource.
func ListAssignments(client *gophercloud.ServiceClient, opts ListAssignmentsOptsBuilder) pagination.Pager {
	url := listAssignmentsURL(client)
	if opts != nil {
		query, err := opts.ToRolesListAssignmentsQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}
	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return RoleAssignmentPage{pagination.LinkedPageBase{PageResult: r}}
	})
}
