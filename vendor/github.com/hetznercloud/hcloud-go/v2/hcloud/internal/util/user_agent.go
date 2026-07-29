package util

func BuildUserAgent(name, version, userAgent string) string {
	switch {
	case name != "" && version != "":
		return name + "/" + version + " " + userAgent
	case name != "" && version == "":
		return name + " " + userAgent
	default:
		return userAgent
	}
}
