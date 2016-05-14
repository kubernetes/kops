package fi

// HasAddress is implemented by elastic/floating IP addresses, to expose the address
// For example, this is used so that the master SSL certificate can be configured with the dynamically allocated IP
type HasAddress interface {
	// FindAddress returns the address associated with the implementor.  If there is no address, returns (nil, nil)
	FindAddress(context *Context) (*string, error)
}
