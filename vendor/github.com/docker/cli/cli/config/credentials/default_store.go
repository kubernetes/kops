package credentials

import "os/exec"

// DetectDefaultStore returns the credentials store to use if no user-defined
// custom helper is passed.
//
// Some platforms define a preferred helper, in which case it attempts to look
// up the helper binary before falling back to the platform's default.
//
// If no user-defined helper is passed, and no helper is found, it returns an
// empty string, which means credentials are stored unencrypted in the CLI's
// config-file without the use of a credentials store.
func DetectDefaultStore(customStore string) string {
	if customStore != "" {
		// use user-defined
		return customStore
	}

	platformDefault := defaultCredentialsStore()
	if platformDefault == "" {
		return ""
	}

	if _, err := exec.LookPath(remoteCredentialsPrefix + platformDefault); err != nil {
		return ""
	}
	return platformDefault
}
