//go:build !amd64 || gccgo
// +build !amd64 gccgo

package abi

func init() {
	cpuid = func(uint32) (uint32, uint32, uint32, uint32) {
		return 0, 0, 0, 0
	}
}
