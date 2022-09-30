// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Vladimir Skipor <skipor@yandex-team.ru>

package operation

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Copy from bb.yandex-team.ru/cloud/cloud-go/pkg/protoutil/any.go
func UnmarshalAny(msg *anypb.Any) (proto.Message, error) {
	if msg == nil {
		return nil, nil
	}
	return msg.UnmarshalNew()
}
