package linodego

import (
	"context"
)

// ObjectStorageTransfer is an object matching the response of object-storage/transfer
type ObjectStorageTransfer struct {
	AmmountUsed int `json:"used"`
}

// CancelObjectStorage cancels and removes all object storage from the Account
func (c *Client) CancelObjectStorage(ctx context.Context) error {
	return doPOSTRequestNoRequestResponseBody(ctx, c, "object-storage/cancel")
}

// GetObjectStorageTransfer returns the amount of outbound data transferred used by the Account
func (c *Client) GetObjectStorageTransfer(ctx context.Context) (*ObjectStorageTransfer, error) {
	return doGETRequest[ObjectStorageTransfer](ctx, c, "object-storage/transfer")
}
