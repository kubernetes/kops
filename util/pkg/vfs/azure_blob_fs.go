package vfs

type AzureBlobPath struct {
	container string
	key       string
}

func newAzureBlobPath(azureBlobCtx *AzureBlobContext, container string, key string) *AzureBlobPath {
	return &AzureBlobPath{
		container: container,
		key:       key,
	}
}

func (a *AzureBlobPath) Join(relativePath ...string) Path {
	return &AzureBlobPath{}
}

func (a *AzureBlobPath) ReadFile() ([]byte, error) {
	return []byte(""), nil
}

func (a *AzureBlobPath) WriteFile(data []byte) error {
	return nil
}

func (a *AzureBlobPath) CreateFile(data []byte) error {
	return nil
}

func (a *AzureBlobPath) Remove() error {
	return nil
}
func (a *AzureBlobPath) Base() string {
	return ""
}
func (a *AzureBlobPath) Path() string {
	return ""
}
func (a *AzureBlobPath) ReadDir() ([]Path, error) {
	var paths []Path
	//paths = append(paths, AzureBlobPath{})
	return paths, nil
}
func (a *AzureBlobPath) ReadTree() ([]Path, error) {
	var paths []Path
	//paths = append(paths, AzureBlobPath{})
	return paths, nil
}
