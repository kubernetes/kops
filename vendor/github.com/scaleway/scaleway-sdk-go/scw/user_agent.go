package scw

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
)

const (
	UserAgentMaxLength = 512
	TruncatedSuffix    = "...[truncated]"
)

var (
	// Default user agent used by NewClient()
	defaultUserAgent = fmt.Sprintf("scaleway-sdk-go/%s (%s; %s; %s)", getVersion(), runtime.Version(), runtime.GOOS, runtime.GOARCH)

	// Standard printable ASCII regex: space (0x20) through tilde (0x7E).
	// This explicitly blocks \r, \n, \t, null bytes (\0), and non-printable control codes.
	printableASCIIRegex = regexp.MustCompile(`^[ -~]+$`)
)

// SanitizeForLogging strips newlines, tabs, and limits length for safe
// insertion into log streams.
func SanitizeForLogging(ua string) string {
	// Truncate to max length if necessary
	if len(ua) > UserAgentMaxLength {
		ua = ua[:UserAgentMaxLength] + TruncatedSuffix
	}

	// Neutralize potential log injection characters
	replacer := strings.NewReplacer(
		"\r", "\\r",
		"\n", "\\n",
		"\t", "\\t",
	)

	return replacer.Replace(ua)
}
