package hcloud

// Architecture specifies the architecture of the CPU.
type Architecture string

const (
	// ArchitectureX86 is the architecture for Intel/AMD x86 CPUs.
	ArchitectureX86 Architecture = "x86"

	// ArchitectureARM is the architecture for ARM CPUs.
	ArchitectureARM Architecture = "arm"
)
