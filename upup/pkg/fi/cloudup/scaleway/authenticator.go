package scaleway

import (
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"k8s.io/kops/pkg/bootstrap"
)

const ScalewayAuthenticationTokenPrefix = "x-scaleway-instance-server-id "

type scalewayAuthenticator struct{}

var _ bootstrap.Authenticator = &scalewayAuthenticator{}

func NewScalewayAuthenticator() (bootstrap.Authenticator, error) {
	return &scalewayAuthenticator{}, nil
}

func (a *scalewayAuthenticator) CreateToken(body []byte) (string, error) {
	metadataAPI := instance.NewMetadataAPI()
	metadata, err := metadataAPI.GetMetadata()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve server metadata: %w", err)
	}
	return ScalewayAuthenticationTokenPrefix + metadata.ID, nil
}
