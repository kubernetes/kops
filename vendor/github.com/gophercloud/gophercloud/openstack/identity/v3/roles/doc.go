/*
Package roles provides information and interaction with the roles API
resource for the OpenStack Identity service.

Example to List Role Assignments

	listOpts := roles.ListAssignmentsOpts{
		UserID:         "97061de2ed0647b28a393c36ab584f39",
		ScopeProjectID: "9df1a02f5eb2416a9781e8b0c022d3ae",
	}

	allPages, err := roles.ListAssignments(identityClient, listOpts).AllPages()
	if err != nil {
		panic(err)
	}

	allRoles, err := roles.ExtractRoleAssignments(allPages)
	if err != nil {
		panic(err)
	}

	for _, role := range allRoles {
		fmt.Printf("%+v\n", role)
	}
*/
package roles
