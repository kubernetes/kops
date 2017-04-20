package view

import (
	"context"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
)

type ContainerView struct {
	object.Common
}

func NewContainerView(c *vim25.Client, ref types.ManagedObjectReference) *ContainerView {
	return &ContainerView{
		Common: object.NewCommon(c, ref),
	}
}

func (v ContainerView) Destroy(ctx context.Context) error {
	req := types.DestroyView{
		This: v.Reference(),
	}
	_, err := methods.DestroyView(ctx, v.Client(), &req)
	return err
}
