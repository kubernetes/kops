package vfs

import "path"

// https://kopsdevel.blob.core.windows.net
type AzureBlobPath struct {
	azureBlobContext *AzureBlobContext
	container        string
	key              string
}

func newAzureBlobPath(azureBlobCtx *AzureBlobContext, container string, key string) *AzureBlobPath {
	return &AzureBlobPath{
		container:        container,
		key:              key,
		azureBlobContext: azureBlobCtx,
	}
}

func (a *AzureBlobPath) Join (relativePath ...string) Path {
	args := []string{a.key}
	args = append(args, relativePath...)
	joined := path.Join(args...)
	return &AzureBlobPath{
		azureBlobContext: a.azureBlobContext,
		container:        a.container,
		key:              joined,
	}
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
	return "https://" + a.container + ".blob.core.windows.net/" + a.container + "/" + a.key
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
