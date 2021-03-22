package deprecations

import "k8s.io/klog"

type Deprecation struct {
	Key              string
	DeprecatedInKops string
}

// ShouldIssueWithCSRs is the deprecation to use CertificateSigningRequests instead of Certs in BootstrapRequest.
// CSRs demonstrate ownership of the public key.
var ShouldIssueWithCSRs = Deprecation{Key: "ShouldIssueWithCSRs", DeprecatedInKops: "1.21"}

func (d *Deprecation) Use() {
	klog.Warningf("using deprecated functionality %q", d.Key)
}

func (d *Deprecation) IsEnabled() bool {
	klog.Infof("continuing to use deprecated codepath for %q", d.Key)
	return true
}
