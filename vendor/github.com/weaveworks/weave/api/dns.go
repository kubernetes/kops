package api

import (
	"fmt"
	"net/url"
)

func (client *Client) DNSDomain() (string, error) {
	return client.httpVerb("GET", "/domain", nil)
}

func (client *Client) RegisterWithDNS(ID string, fqdn string, ip string) error {
	data := url.Values{}
	data.Add("fqdn", fqdn)
	_, err := client.httpVerb("PUT", fmt.Sprintf("/name/%s/%s", ID, ip), data)
	return err
}

func (client *Client) DeregisterWithDNS(ID string, ip string) error {
	_, err := client.httpVerb("DELETE", fmt.Sprintf("/name/%s/%s", ID, ip), nil)
	return err
}
