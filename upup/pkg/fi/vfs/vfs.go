package vfs

// Yet another VFS package
// If there's a "winning" VFS implementation in go, we should switch to it!

type VFS interface {
}

type Path interface {
	Join(relativePath ...string) Path
	ReadFile() ([]byte, error)

	WriteFile(data []byte) error
	// CreateFile writes the file contents, but only if the file does not already exist
	CreateFile(data []byte) error

	// Base returns the base name (last element)
	Base() string

	ReadDir() ([]Path, error)
}
