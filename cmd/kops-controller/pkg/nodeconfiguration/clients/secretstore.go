package clients

import (
	"fmt"

	pb "k8s.io/kops/pkg/proto/nodeconfiguration"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

//grpcSecretStore is a SecretStore backed by the GRPC client
type grpcSecretStore struct {
	client     pb.NodeConfigurationServiceClient
	nodeConfig pb.GetConfigurationResponse
}

func NewSecretStore(client pb.NodeConfigurationServiceClient, nodeConfig *pb.GetConfigurationResponse) fi.SecretStore {
	return &grpcSecretStore{
		client:     client,
		nodeConfig: *nodeConfig,
	}
}

// Secret implements fi.SecretStore
func (s *grpcSecretStore) Secret(id string) (*fi.Secret, error) {
	return nil, fmt.Errorf("Secret not supported by grpcSecretStore")
}

// DeleteSecret implements fi.SecretStore
func (s *grpcSecretStore) DeleteSecret(id string) error {
	return fmt.Errorf("DeleteSecret not supported by grpcSecretStore")
}

// FindSecret implements fi.SecretStore
func (s *grpcSecretStore) FindSecret(id string) (*fi.Secret, error) {
	return nil, fmt.Errorf("FindSecret not supported by grpcSecretStore")
}

// GetOrCreateSecret implements fi.SecretStore
func (s *grpcSecretStore) GetOrCreateSecret(id string, secret *fi.Secret) (current *fi.Secret, created bool, err error) {
	return nil, false, fmt.Errorf("GetOrCreateSecret not supported by grpcSecretStore")
}

// ReplaceSecret implements fi.SecretStore
func (s *grpcSecretStore) ReplaceSecret(id string, secret *fi.Secret) (current *fi.Secret, err error) {
	return nil, fmt.Errorf("ReplaceSecret not supported by grpcSecretStore")
}

// ListSecrets implements fi.SecretStore
func (s *grpcSecretStore) ListSecrets() ([]string, error) {
	return nil, fmt.Errorf("ListSecrets not supported by grpcSecretStore")
}

// MirrorTo implements fi.SecretStore
func (s *grpcSecretStore) MirrorTo(basedir vfs.Path) error {
	return fmt.Errorf("MirrorTo not supported by grpcSecretStore")
}
