package hcloud

import (
	"context"
	"fmt"
	"net"
)

// RDNSSupporter defines functions to change and lookup reverse dns entries.
// currently implemented by Server, FloatingIP and LoadBalancer.
type RDNSSupporter interface {
	// changeDNSPtr changes or resets the reverse DNS pointer for a IP address.
	// Pass a nil ptr to reset the reverse DNS pointer to its default value.
	changeDNSPtr(ctx context.Context, client *Client, ip net.IP, ptr *string) (*Action, *Response, error)
	// GetDNSPtrForIP searches for the dns assigned to the given IP address.
	// It returns an error if there is no dns set for the given IP address.
	GetDNSPtrForIP(ip net.IP) (string, error)
}

// RDNSClient simplifys the handling objects which support reverse dns entries.
type RDNSClient struct {
	client *Client
}

// ChangeDNSPtr changes or resets the reverse DNS pointer for a IP address.
// Pass a nil ptr to reset the reverse DNS pointer to its default value.
func (c *RDNSClient) ChangeDNSPtr(ctx context.Context, rdns RDNSSupporter, ip net.IP, ptr *string) (*Action, *Response, error) {
	return rdns.changeDNSPtr(ctx, c.client, ip, ptr)
}

// SupportsRDNS checks if the object supports reverse dns functions.
func SupportsRDNS(i interface{}) bool {
	_, ok := i.(RDNSSupporter)
	return ok
}

// RDNSLookup searches for the dns assigned to the given IP address.
// It returns an error if the object does not support reverse dns or if there is no dns set for the given IP address.
func RDNSLookup(i interface{}, ip net.IP) (string, error) {
	rdns, ok := i.(RDNSSupporter)
	if !ok {
		return "", fmt.Errorf("%+v does not support RDNS", i)
	}

	return rdns.GetDNSPtrForIP(ip)
}
