package util

import (
	"k8s.io/client-go/pkg/api/v1"
	"strings"
)

func GetNodeRole(node *v1.Node) string {
	role := ""
	// Newer labels
	for k := range node.Labels {
		if strings.HasPrefix(k, "node-role.kubernetes.io/") {
			role = strings.TrimPrefix(k, "node-role.kubernetes.io/")
		}
	}
	// Older label
	if role == "" {
		role = node.Labels["kubernetes.io/role"]
	}

	return role
}
