package ycsdk

import (
	"github.com/yandex-cloud/go-sdk/gen/cdn"
)

const (
	CDNID = "cdn"
)

func (sdk *SDK) CDN() *cdn.CDN {
	return cdn.NewCDN(sdk.getConn(CDNID))
}
