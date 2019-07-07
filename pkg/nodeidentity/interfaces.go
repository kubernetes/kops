package nodeidentity

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

type Identifier interface {
	IdentifyNode(ctx context.Context, node *corev1.Node) (*Info, error)
}

type Info struct {
	InstanceGroup string
}
