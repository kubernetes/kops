// Copyright (c) 2019 YANDEX LLC.

package ycsdk

import (
	"github.com/yandex-cloud/go-sdk/gen/apploadbalancer"
)

const (
	ApplicationLoadBalancerServiceID Endpoint = "alb"
)

// LoadBalancer returns LoadBalancer object that is used to operate on load balancers
func (sdk *SDK) ApplicationLoadBalancer() *apploadbalancer.LoadBalancer {
	return apploadbalancer.NewLoadBalancer(sdk.getConn(ApplicationLoadBalancerServiceID))
}
