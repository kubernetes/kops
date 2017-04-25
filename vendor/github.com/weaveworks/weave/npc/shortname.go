package npc

import (
	"crypto/sha1"
	"math/big"
)

const (
	// This array:
	// * Must include only printable UTF8 characters that are represented with a single byte
	// * Must be at least of length 85 (`len("weave-") + l(2^160)/l(85)` equals 31, the maximum ipset name length)
	// * Must not include commas as those are treated specially by `ipset add` when adding a named set to a list:set
	// * Should not include space for readability
	// * Should not include single quote or backslash to be nice to shell users
	ShortNameSymbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789(){}[]<>_$%^&*|/?.;:@#~"
)

// sha1 hash an arbitrary string and represent it using the full range of
// printable ascii characters
func shortName(arbitrary string) string {
	symbols := []byte(ShortNameSymbols)
	sum := sha1.Sum([]byte(arbitrary))
	i := big.NewInt(0).SetBytes(sum[:])
	base := big.NewInt(int64(len(symbols)))
	zero := big.NewInt(0)
	result := make([]byte, 0)

	for i.Cmp(zero) > 0 {
		remainder := new(big.Int).Mod(i, base)
		i.Sub(i, remainder)
		i.Div(i, base)
		result = append(result, symbols[remainder.Int64()])
	}

	return string(result)
}
