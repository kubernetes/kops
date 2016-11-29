package dns

import "testing"

func TestAliasForNodesInRole(t *testing.T) {
	cases := []struct {
		role, roleType, expected string
	}{
		{"node", RoleTypeExternal, "node/role=node/external"},
		{"node", RoleTypeInternal, "node/role=node/internal"},
	}

	for _, c := range cases {
		if actual := AliasForNodesInRole(c.role, c.roleType); actual != c.expected {
			t.Errorf("AliasForNodesInRole(%#v, %#v) expected %#v, but got %#v", c.role, c.roleType, c.expected, actual)
		}
	}
}
