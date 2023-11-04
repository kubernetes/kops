//go:build amd64 && !gccgo
// +build amd64,!gccgo

package abi

func asmCpuid(op uint32) (eax, ebx, ecx, edx uint32)

func init() {
	cpuid = asmCpuid
}
