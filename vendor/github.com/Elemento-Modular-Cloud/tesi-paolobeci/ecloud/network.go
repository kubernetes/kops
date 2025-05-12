package ecloud

import (
	"context"
)

// Network represents a network in Elemento Cloud.
type Network struct {
	ID          string
	Name        string
	Description string
	Created     string
	Updated     string
	Labels      map[string]string
	Region      string
	IPRange     string
	Subnets	    []Subnet
}

// Subnet represents a subnet in Elemento Cloud.
type Subnet struct {
	IPRange     string
	NetworkZone string
}

// NetworkClient is a client for the servers API.
type NetworkClient struct {
	client *Client
}

// Get Network returns the network with the given ID.
func (c *NetworkClient) Get(ctx context.Context, id string) (*Network, error) {
	// TODO: da creare le funzioni per la network
	// network := &Network{}
	// err := c.client.get("network", id, network)
	// if err != nil {
	// 	return nil, err
	// }
	// return network, nil
	
	return nil, nil
}