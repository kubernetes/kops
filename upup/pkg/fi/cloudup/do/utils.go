package do

import "strings"

func SafeClusterName(clusterName string) string {
	// DO does not support . in tags / names
	safeClusterName := strings.Replace(clusterName, ".", "-", -1)
	return safeClusterName
}
