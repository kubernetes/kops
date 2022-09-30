package ycsdk

import (
	lockboxpayload "github.com/yandex-cloud/go-sdk/gen/lockboxpayload"
	lockboxsecret "github.com/yandex-cloud/go-sdk/gen/lockboxsecret"
)

const (
	LockboxSecretServiceID  = "lockbox"
	LockboxPayloadServiceID = "lockbox-payload"
)

func (sdk *SDK) LockboxSecret() *lockboxsecret.LockboxSecret {
	return lockboxsecret.NewLockboxSecret(sdk.getConn(LockboxSecretServiceID))
}

func (sdk *SDK) LockboxPayload() *lockboxpayload.LockboxPayload {
	return lockboxpayload.NewLockboxPayload(sdk.getConn(LockboxPayloadServiceID))
}
