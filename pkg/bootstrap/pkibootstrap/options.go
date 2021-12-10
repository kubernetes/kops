package pkibootstrap

// Options describes how we authenticate instances with GCE TPM authentication.
type Options struct {
	// MaxTimeSkew is the maximum time skew to allow (in seconds)
	MaxTimeSkew int64 `json:"MaxTimeSkew,omitempty"`
}

// AuthenticationTokenPrefix is the prefix used for authentication using PKI
const AuthenticationTokenPrefix = "x-pki-tpm "
