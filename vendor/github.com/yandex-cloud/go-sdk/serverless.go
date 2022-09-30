package ycsdk

import (
	"github.com/yandex-cloud/go-sdk/gen/apigateway"
	"github.com/yandex-cloud/go-sdk/gen/apigateway/websocket"
	"github.com/yandex-cloud/go-sdk/gen/containers"
	"github.com/yandex-cloud/go-sdk/gen/functions"
	"github.com/yandex-cloud/go-sdk/gen/mdbproxy"
	"github.com/yandex-cloud/go-sdk/gen/triggers"
)

type Serverless struct {
	sdk *SDK
}

const (
	FunctionServiceID             Endpoint = "serverless-functions"
	TriggerServiceID              Endpoint = "serverless-triggers"
	APIGatewayServiceID           Endpoint = "serverless-apigateway"
	MDBProxyServiceID             Endpoint = "mdbproxy"
	ServerlessContainersServiceID Endpoint = "serverless-containers"
	APIGatewayWebsocketServiceID  Endpoint = "apigateway-connections"
)

func (s *Serverless) Functions() *functions.Function {
	return functions.NewFunction(s.sdk.getConn(FunctionServiceID))
}

func (s *Serverless) Triggers() *triggers.Trigger {
	return triggers.NewTrigger(s.sdk.getConn(TriggerServiceID))
}

func (s *Serverless) APIGateway() *apigateway.Apigateway {
	return apigateway.NewApigateway(s.sdk.getConn(APIGatewayServiceID))
}

func (s *Serverless) MDBProxy() *mdbproxy.Proxy {
	return mdbproxy.NewProxy(s.sdk.getConn(MDBProxyServiceID))
}

func (s *Serverless) Containers() *containers.Container {
	return containers.NewContainer(s.sdk.getConn(ServerlessContainersServiceID))
}

func (s *Serverless) APIGatewayWebsocket() *websocket.Websocket {
	return websocket.NewWebsocket(s.sdk.getConn(APIGatewayWebsocketServiceID))
}
