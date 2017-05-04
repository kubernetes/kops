package vfs

import (
	"github.com/Azure/azure-sdk-for-go/storage"
	"os"
	"github.com/pkg/errors"
	"fmt"
	"sync"
)

type AzureBlobContext struct {
	mutex  sync.Mutex
	client storage.Client
}

func NewAzureBlobContext() *AzureBlobContext {
	return &AzureBlobContext{}
}

// getClient holds the primary connection logic for connecting to the Azure
// API via the Azure Go SDK
func (a *AzureBlobContext) getClient() (storage.Client, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.client == nil {
		// Now we assume that they are defined as environmental variables
		name := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
		if name == "" {
			return nil, errors.New("invalid or empty value for $AZURE_STORAGE_ACCOUNT_NAME")
		}
		key := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
		if key == "" {
			return nil, errors.New("invalid or empty value for $AZURE_STORAGE_ACCOUNT_KEY")
		}

		client, err := storage.NewBasicClient(name, key)
		if err != nil {
			return nil, fmt.Errorf("unable to create Azure connection: %v", err)
		}
		a.client = client
	}
	return a.client, nil
}
