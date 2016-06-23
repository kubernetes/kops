package fi

import (
	"bytes"
	crypto_rand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi/vfs"
	"math/big"
	"os"
	"strings"
	"time"
)

type VFSCAStore struct {
	dryrun         bool
	basedir        vfs.Path
	caCertificates *certificates
	caPrivateKeys  *privateKeys
}

var _ CAStore = &VFSCAStore{}

func NewVFSCAStore(basedir vfs.Path, dryrun bool) (CAStore, error) {
	c := &VFSCAStore{
		dryrun:  dryrun,
		basedir: basedir,
	}
	//err := os.MkdirAll(path.Join(basedir, "private"), 0700)
	//if err != nil {
	//	return nil, fmt.Errorf("error creating directory: %v", err)
	//}
	//err = os.MkdirAll(path.Join(basedir, "issued"), 0700)
	//if err != nil {
	//	return nil, fmt.Errorf("error creating directory: %v", err)
	//}
	caCertificates, err := c.loadCertificates(c.buildCertificatePoolPath(CertificateId_CA))
	if err != nil {
		return nil, err
	}

	if caCertificates != nil {
		caPrivateKeys, err := c.loadPrivateKeys(c.buildPrivateKeyPoolPath(CertificateId_CA))
		if err != nil {
			return nil, err
		}
		if caPrivateKeys == nil {
			glog.Warningf("CA private key was not found")
			//return nil, fmt.Errorf("error loading CA private key - key not found")
		}
		c.caCertificates = caCertificates
		c.caPrivateKeys = caPrivateKeys
	} else {
		err := c.generateCACertificate()
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (s *VFSCAStore) VFSPath() vfs.Path {
	return s.basedir
}

func (c *VFSCAStore) generateCACertificate() error {
	subject := &pkix.Name{
		CommonName: "kubernetes",
	}
	serial := c.buildSerial()
	template := &x509.Certificate{
		Subject:               *subject,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		BasicConstraintsValid: true,
		IsCA: true,
	}

	caRsaKey, err := rsa.GenerateKey(crypto_rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("error generating RSA private key: %v", err)
	}

	caPrivateKey := &PrivateKey{Key: caRsaKey}

	caCertificate, err := SignNewCertificate(caPrivateKey, template, nil, nil)
	if err != nil {
		return err
	}

	keyPath := c.buildPrivateKeyPath(CertificateId_CA, serial)
	err = c.storePrivateKey(caPrivateKey, keyPath)
	if err != nil {
		return err
	}

	// Make double-sure it round-trips
	privateKeys, err := c.loadPrivateKeys(c.buildPrivateKeyPoolPath(CertificateId_CA))
	if err != nil {
		return err
	}
	if privateKeys == nil || privateKeys.primary != serial.Text(10) {
		return fmt.Errorf("failed to round-trip CA private key")
	}

	certPath := c.buildCertificatePath(CertificateId_CA, serial)
	err = c.storeCertificate(caCertificate, certPath)
	if err != nil {
		return err
	}

	// Make double-sure it round-trips
	certificates, err := c.loadCertificates(c.buildCertificatePoolPath(CertificateId_CA))
	if err != nil {
		return err
	}

	if certificates == nil || certificates.primary != serial.Text(10) {
		return fmt.Errorf("failed to round-trip CA certifiacate")
	}

	c.caPrivateKeys = privateKeys
	c.caCertificates = certificates
	return nil
}

func (c *VFSCAStore) buildCertificatePoolPath(id string) vfs.Path {
	return c.basedir.Join("issued", id)
}

func (c *VFSCAStore) buildCertificatePath(id string, serial *big.Int) vfs.Path {
	return c.basedir.Join("issued", id, serial.Text(10)+".crt")
}

func (c *VFSCAStore) buildPrivateKeyPoolPath(id string) vfs.Path {
	return c.basedir.Join("private", id)
}

func (c *VFSCAStore) buildPrivateKeyPath(id string, serial *big.Int) vfs.Path {
	return c.basedir.Join("private", id, serial.Text(10)+".key")
}

type certificates struct {
	certificates map[string]*Certificate
	primary      string
}

func (p *certificates) Primary() *Certificate {
	if p.primary == "" {
		return nil
	}
	return p.certificates[p.primary]
}

func (c *VFSCAStore) loadCertificates(p vfs.Path) (*certificates, error) {
	files, err := p.ReadDir()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	certs := &certificates{
		certificates: make(map[string]*Certificate),
	}

	for _, f := range files {
		cert, err := c.loadOneCertificate(f)
		if err != nil {
			return nil, fmt.Errorf("error loading certificate %q: %v", f, err)
		}
		name := f.Base()
		name = strings.TrimSuffix(name, ".crt")
		certs.certificates[name] = cert
	}

	if len(certs.certificates) == 0 {
		return nil, nil
	}

	var primaryVersion *big.Int
	for k := range certs.certificates {
		version, ok := big.NewInt(0).SetString(k, 10)
		if !ok {
			glog.Warningf("Ignoring certificate with non-integer version: %q", k)
			continue
		}

		if primaryVersion == nil || version.Cmp(primaryVersion) > 0 {
			certs.primary = k
			primaryVersion = version
		}
	}

	return certs, nil
}

func (c *VFSCAStore) loadOneCertificate(p vfs.Path) (*Certificate, error) {
	data, err := p.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	cert, err := LoadPEMCertificate(data)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return nil, nil
	}
	return cert, nil
}

func (c *VFSCAStore) Cert(id string) (*Certificate, error) {
	cert, err := c.FindCert(id)
	if err == nil && cert == nil {
		if c.dryrun {
			glog.Warningf("using empty certificate, because --dryrun specified")
			return &Certificate{}, err
		}
		return nil, fmt.Errorf("cannot find certificate %q", id)
	}
	return cert, err

}

func (c *VFSCAStore) CertificatePool(id string) (*CertificatePool, error) {
	cert, err := c.FindCertificatePool(id)
	if err == nil && cert == nil {
		if c.dryrun {
			glog.Warningf("using empty certificate, because --dryrun specified")
			return &CertificatePool{}, err
		}
		return nil, fmt.Errorf("cannot find certificate pool %q", id)
	}
	return cert, err

}

func (c *VFSCAStore) FindCert(id string) (*Certificate, error) {
	var certs *certificates

	if id == CertificateId_CA {
		certs = c.caCertificates
	} else {
		var err error
		p := c.buildCertificatePoolPath(id)
		certs, err = c.loadCertificates(p)
		if err != nil {
			return nil, err
		}
	}

	var cert *Certificate
	if certs != nil && certs.primary != "" {
		cert = certs.certificates[certs.primary]
	}

	return cert, nil
}

func (c *VFSCAStore) FindCertificatePool(id string) (*CertificatePool, error) {
	var certs *certificates

	if id == CertificateId_CA {
		certs = c.caCertificates
	} else {
		var err error
		p := c.buildCertificatePoolPath(id)
		certs, err = c.loadCertificates(p)
		if err != nil {
			return nil, err
		}
	}

	pool := &CertificatePool{}

	if certs != nil {
		pool.Primary = certs.Primary()

		for k, cert := range certs.certificates {
			if k == certs.primary {
				continue
			}
			pool.Secondary = append(pool.Secondary, cert)
		}
	}
	return pool, nil
}

func (c *VFSCAStore) List() ([]string, error) {
	var ids []string

	issuedDir := c.basedir.Join("issued")
	files, err := issuedDir.ReadDir()
	if err != nil {
		return nil, fmt.Errorf("error reading directory %q: %v", issuedDir, err)
	}

	for _, f := range files {
		name := f.Base()
		ids = append(ids, name)
	}
	return ids, nil
}

func (c *VFSCAStore) IssueCert(id string, serial *big.Int, privateKey *PrivateKey, template *x509.Certificate) (*Certificate, error) {
	glog.Infof("Issuing new certificate: %q", id)

	template.SerialNumber = serial

	p := c.buildCertificatePath(id, serial)

	if c.caPrivateKeys == nil || c.caPrivateKeys.Primary() == nil {
		return nil, fmt.Errorf("ca.key was not found; cannot issue certificates")
	}
	cert, err := SignNewCertificate(privateKey, template, c.caCertificates.Primary().Certificate, c.caPrivateKeys.Primary())
	if err != nil {
		return nil, err
	}

	err = c.storeCertificate(cert, p)
	if err != nil {
		return nil, err
	}

	// Make double-sure it round-trips
	return c.loadOneCertificate(p)
}

func (c *VFSCAStore) AddCert(id string, cert *Certificate) error {
	glog.Infof("Issuing new certificate: %q", id)

	// We add with a timestamp of zero so this will never be the newest cert
	serial := buildSerial(0)

	p := c.buildCertificatePath(id, serial)

	err := c.storeCertificate(cert, p)
	if err != nil {
		return err
	}

	// Make double-sure it round-trips
	_, err = c.loadOneCertificate(p)
	return err
}

type privateKeys struct {
	keys    map[string]*PrivateKey
	primary string
}

func (p *privateKeys) Primary() *PrivateKey {
	if p.primary == "" {
		return nil
	}
	return p.keys[p.primary]
}

func (c *VFSCAStore) loadPrivateKeys(p vfs.Path) (*privateKeys, error) {
	files, err := p.ReadDir()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	keys := &privateKeys{
		keys: make(map[string]*PrivateKey),
	}

	for _, f := range files {
		key, err := c.loadOnePrivateKey(f)
		if err != nil {
			return nil, fmt.Errorf("error loading private key %q: %v", f, err)
		}
		name := f.Base()
		name = strings.TrimSuffix(name, ".key")
		keys.keys[name] = key
	}

	if len(keys.keys) == 0 {
		return nil, nil
	}

	var primaryVersion *big.Int
	for k := range keys.keys {
		version, ok := big.NewInt(0).SetString(k, 10)
		if !ok {
			glog.Warningf("Ignoring private key with non-integer version: %q", k)
			continue
		}

		if primaryVersion == nil || version.Cmp(primaryVersion) > 0 {
			keys.primary = k
			primaryVersion = version
		}
	}

	return keys, nil
}

func (c *VFSCAStore) loadOnePrivateKey(p vfs.Path) (*PrivateKey, error) {
	data, err := p.ReadFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	k, err := ParsePEMPrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key from %q: %v", p, err)
	}
	return k, err
}

func ParsePEMPrivateKey(data []byte) (*PrivateKey, error) {
	k, err := parsePEMPrivateKey(data)
	if err != nil {
		return nil, err
	}
	if k == nil {
		return nil, nil
	}
	return &PrivateKey{Key: k}, nil
}

func (c *VFSCAStore) FindPrivateKey(id string) (*PrivateKey, error) {
	var keys *privateKeys
	if id == CertificateId_CA {
		keys = c.caPrivateKeys
	} else {
		var err error
		p := c.buildPrivateKeyPoolPath(id)
		keys, err = c.loadPrivateKeys(p)
		if err != nil {
			return nil, err
		}

	}

	var key *PrivateKey
	if keys != nil && keys.primary != "" {
		key = keys.keys[keys.primary]
	}
	return key, nil
}

func (c *VFSCAStore) PrivateKey(id string) (*PrivateKey, error) {
	key, err := c.FindPrivateKey(id)
	if err == nil && key == nil {
		if c.dryrun {
			glog.Warningf("using empty certificate, because --dryrun specified")
			return &PrivateKey{}, err
		}
		return nil, fmt.Errorf("cannot find SSL key %q", id)
	}
	return key, err

}

func (c *VFSCAStore) CreateKeypair(id string, template *x509.Certificate) (*Certificate, *PrivateKey, error) {
	serial := c.buildSerial()

	privateKey, err := c.CreatePrivateKey(id, serial)
	if err != nil {
		return nil, nil, err
	}

	cert, err := c.IssueCert(id, serial, privateKey, template)
	if err != nil {
		// TODO: Delete cert?
		return nil, nil, err
	}

	return cert, privateKey, nil
}

func (c *VFSCAStore) CreatePrivateKey(id string, serial *big.Int) (*PrivateKey, error) {
	p := c.buildPrivateKeyPath(id, serial)

	rsaKey, err := rsa.GenerateKey(crypto_rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("error generating RSA private key: %v", err)
	}

	privateKey := &PrivateKey{Key: rsaKey}
	err = c.storePrivateKey(privateKey, p)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func (c *VFSCAStore) storePrivateKey(privateKey *PrivateKey, p vfs.Path) error {
	var data bytes.Buffer
	_, err := privateKey.WriteTo(&data)
	if err != nil {
		return err
	}

	return p.WriteFile(data.Bytes())
}

func (c *VFSCAStore) storeCertificate(cert *Certificate, p vfs.Path) error {
	// TODO: replace storePrivateKey & storeCertificate with writeFile(io.WriterTo)?
	var data bytes.Buffer
	_, err := cert.WriteTo(&data)
	if err != nil {
		return err
	}

	return p.WriteFile(data.Bytes())
}

func (c *VFSCAStore) buildSerial() *big.Int {
	t := time.Now().UnixNano()
	return buildSerial(t)
}

func buildSerial(timestamp int64) *big.Int {
	randomLimit := new(big.Int).Lsh(big.NewInt(1), 32)
	randomComponent, err := crypto_rand.Int(crypto_rand.Reader, randomLimit)
	if err != nil {
		glog.Fatalf("error generating random number: %v", err)
	}

	serial := big.NewInt(timestamp)
	serial.Lsh(serial, 32)
	serial.Or(serial, randomComponent)

	return serial
}
