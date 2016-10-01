package nodeup

type ProtokubeFlags struct {
	DNSZoneName   *string `json:"dnsZoneName,omitempty" flag:"dns-zone-name"`
	Master        *bool   `json:"master,omitempty" flag:"master"`
	Containerized *bool   `json:"containerized,omitempty" flag:"containerized"`
	LogLevel      *int    `json:"logLevel,omitempty" flag:"v"`

	Channels []string `json:"channels,omitempty" flag:"channels"`
}
