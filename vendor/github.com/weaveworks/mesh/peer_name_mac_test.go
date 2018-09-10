// +build peer_name_mac !peer_name_alternative

package mesh_test

import (
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/mesh"
	"testing"
)

func TestMacPeerNameFromUserInput(t *testing.T) {
	t.Skip("TODO")
}

func checkSuccess(t *testing.T, nameStr string, expected uint64) {
	actual, err := mesh.PeerNameFromString(nameStr)
	require.NoError(t, err)
	require.Equal(t, mesh.PeerName(expected), actual)
}

func checkFailure(t *testing.T, nameStr string) {
	_, err := mesh.PeerNameFromString(nameStr)
	require.Error(t, err)
}

func TestMacPeerNameFromString(t *testing.T) {
	// Permitted elisions
	checkSuccess(t, "12:34:56:78:9A:BC", 0x123456789ABC)
	checkSuccess(t, "::56:78:9A:BC", 0x000056789ABC)
	checkSuccess(t, "12::78:9A:BC", 0x120000789ABC)
	checkSuccess(t, "12:34::9A:BC", 0x123400009ABC)
	checkSuccess(t, "12:34:56::BC", 0x1234560000BC)
	checkSuccess(t, "12:34:56:78::", 0x123456780000)
	checkSuccess(t, "::78:9A:BC", 0x000000789ABC)
	checkSuccess(t, "12::9A:BC", 0x120000009ABC)
	checkSuccess(t, "12:34::BC", 0x1234000000BC)
	checkSuccess(t, "12:34:56::", 0x123456000000)
	checkSuccess(t, "::9A:BC", 0x000000009ABC)
	checkSuccess(t, "12::BC", 0x1200000000BC)
	checkSuccess(t, "12:34::", 0x123400000000)
	checkSuccess(t, "::BC", 0x0000000000BC)
	checkSuccess(t, "12::", 0x120000000000)

	// Case insensitivity
	checkSuccess(t, "ab:cD:Ef:AB::", 0xABCDEFAB0000)

	// Optional zero padding
	checkSuccess(t, "1:2:3:4:5:6", 0x010203040506)
	checkSuccess(t, "01:02:03:04:05:06", 0x010203040506)

	// Trailing garbage detection
	checkFailure(t, "12::garbage")

	// Octet length
	checkFailure(t, "123::")

	// Forbidden elisions
	checkFailure(t, "::")
	checkFailure(t, "::34:56:78:9A:BC")
	checkFailure(t, "12::56:78:9A:BC")
	checkFailure(t, "12:34::78:9A:BC")
	checkFailure(t, "12:34:56::9A:BC")
	checkFailure(t, "12:34:56:78::BC")
	checkFailure(t, "12:34:56:78:9A::")
	checkFailure(t, "12::78::")
}

func TestMacPeerNameFromBin(t *testing.T) {
	t.Skip("TODO")
}
