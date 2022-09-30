// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Oleg Andriianov <ovandriyanov@yandex-team.ru>

package ycsdk

import (
	"github.com/yandex-cloud/go-sdk/gen/datatransfer"
)

const (
	DataTransferServiceID Endpoint = "datatransfer"
)

// DataTransfer returns DataTransfer object that is used to manage Yandex Data Transfer
func (sdk *SDK) DataTransfer() *datatransfer.DataTransfer {
	return datatransfer.NewDataTransfer(sdk.getConn(DataTransferServiceID))
}
