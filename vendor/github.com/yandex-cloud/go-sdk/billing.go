package ycsdk

import (
	"github.com/yandex-cloud/go-sdk/gen/billing"
)

const (
	BillingServiceID Endpoint = "billing"
)

// Billing returns Billing object that is used to operate on Yandex Billing
func (sdk *SDK) Billing() *billing.Billing {
	return billing.NewBilling(sdk.getConn(BillingServiceID))
}
