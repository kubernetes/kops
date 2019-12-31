package clients

import (
	"context"
	"crypto/x509"
	"fmt"

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/pki"
	pb "k8s.io/kops/pkg/proto/nodeconfiguration"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

//grpcKeyStore is a KeyStore backed by the GRPC client
type grpcKeyStore struct {
	client     pb.NodeConfigurationServiceClient
	nodeConfig pb.GetConfigurationResponse
}

func NewKeyStore(client pb.NodeConfigurationServiceClient, nodeConfig *pb.GetConfigurationResponse) fi.CAStore {
	return &grpcKeyStore{
		client:     client,
		nodeConfig: *nodeConfig,
	}
}

// FindKeypair implements fi.Keystore
func (s *grpcKeyStore) FindKeypair(name string) (*pki.Certificate, *pki.PrivateKey, fi.KeysetFormat, error) {
	ctx := context.Background()
	request := &pb.GetKeypairRequest{
		Name: name,
	}

	klog.V(2).Infof("GetKeypair request to server: %v", request)

	response, err := s.client.GetKeypair(ctx, request)
	if err != nil {
		return nil, nil, "", fmt.Errorf("error fetching keypair %q from server: %v", name, err)
	}

	cert, err := pki.ParsePEMCertificate([]byte(response.Cert))
	if err != nil {
		return nil, nil, "", fmt.Errorf("error parsing certificate: %v", err)
	}

	key, err := pki.ParsePEMPrivateKey([]byte(response.Key))
	if err != nil {
		return nil, nil, "", fmt.Errorf("error parsing key: %v", err)
	}

	return cert, key, "", nil
}

// FindKeypair implements fi.Keystore
func (s *grpcKeyStore) CreateKeypair(signer string, name string, template *x509.Certificate, privateKey *pki.PrivateKey) (*pki.Certificate, error) {
	return nil, fmt.Errorf("CreateKeypair not supported by grpcKeyStore")
}

// FindKeypair implements fi.Keystore
func (s *grpcKeyStore) StoreKeypair(id string, cert *pki.Certificate, privateKey *pki.PrivateKey) error {
	return fmt.Errorf("StoreKeypair not supported by grpcKeyStore")
}

// FindKeypair implements fi.Keystore
func (s *grpcKeyStore) MirrorTo(basedir vfs.Path) error {
	return fmt.Errorf("MirrorTo not supported by grpcKeyStore")
}

// CertificatePool implements fi.CAStore
func (s *grpcKeyStore) CertificatePool(name string, createIfMissing bool) (*fi.CertificatePool, error) {
	return nil, fmt.Errorf("CertificatePool not supported by grpcKeyStore")
}

// FindCertificatePool implements fi.CAStore
func (s *grpcKeyStore) FindCertificatePool(name string) (*fi.CertificatePool, error) {
	return nil, fmt.Errorf("FindCertificatePool not supported by grpcKeyStore")
}

// FindCertificateKeyset implements fi.CAStore
func (s *grpcKeyStore) FindCertificateKeyset(name string) (*kops.Keyset, error) {
	return nil, fmt.Errorf("FindCertificateKeyset not supported by grpcKeyStore")
}

// FindPrivateKey implements fi.CAStore
func (s *grpcKeyStore) FindPrivateKey(name string) (*pki.PrivateKey, error) {
	return nil, fmt.Errorf("FindPrivateKey not supported by grpcKeyStore")
}

// FindPrivateKeyset implements fi.CAStore
func (s *grpcKeyStore) FindPrivateKeyset(name string) (*kops.Keyset, error) {
	return nil, fmt.Errorf("FindPrivateKeyset not supported by grpcKeyStore")
}

// FindCert implements fi.CAStore
func (s *grpcKeyStore) FindCert(name string) (*pki.Certificate, error) {
	if name == "ca" {
		// Special case for the CA certificate
		c, err := pki.ParsePEMCertificate([]byte(s.nodeConfig.CaCertificate))
		if err != nil {
			return nil, fmt.Errorf("error parsing ca certificate: %v", err)
		}
		return c, nil
	}

	return nil, fmt.Errorf("FindCert(%q) not supported by grpcKeyStore", name)
}

// ListKeysets implements fi.CAStore
func (s *grpcKeyStore) ListKeysets() ([]*kops.Keyset, error) {
	return nil, fmt.Errorf("ListKeysets not supported by grpcKeyStore")
}

// AddCert implements fi.CAStore
func (s *grpcKeyStore) AddCert(name string, cert *pki.Certificate) error {
	return fmt.Errorf("AddCert not supported by grpcKeyStore")
}

// DeleteKeysetItem implements fi.CAStore
func (s *grpcKeyStore) DeleteKeysetItem(item *kops.Keyset, id string) error {
	return fmt.Errorf("DeleteKeysetItem not supported by grpcKeyStore")
}
