package protokube

import (
	"fmt"
	"time"
)

const defaultTTL = time.Minute

type DNSProvider interface {
	Set(fqdn string, recordType string, value string, ttl time.Duration) error
}

// MapInternalName maps a FQDN to the internal IP address of the current machine
func (k *KubeBoot) MapInternalDNSName(fqdn string) error {
	err := k.DNS.Set(fqdn, "A", k.InternalIP.String(), defaultTTL)
	if err != nil {
		return fmt.Errorf("error configuring DNS name %q: %v", fqdn, err)
	}
	return nil
}

// BuildInternalDNSName builds a DNS name for use inside the cluster, adding our internal DNS suffix to the key,
func (k *KubeBoot) BuildInternalDNSName(key string) string {
	fqdn := key + k.InternalDNSSuffix
	return fqdn
}
