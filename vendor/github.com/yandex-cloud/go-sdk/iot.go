// Copyright (c) 2019 Yandex LLC. All rights reserved.
// Author: Andrey Khaliullin <avhaliullin@yandex-team.ru>

package ycsdk

import (
	"github.com/yandex-cloud/go-sdk/gen/iot/broker"
	brokerdata "github.com/yandex-cloud/go-sdk/gen/iot/brokerdata"
	data "github.com/yandex-cloud/go-sdk/gen/iot/data"
	"github.com/yandex-cloud/go-sdk/gen/iot/devices"
)

const (
	IoTDevicesServiceID    Endpoint = "iot-devices"
	IoTDataServiceID       Endpoint = "iot-data"
	IoTBrokerServiceID     Endpoint = "iot-broker"
	IoTBrokerDataServiceID Endpoint = "broker-data"
)

func (sdk *SDK) IoT() *IoT {
	return &IoT{sdk: sdk}
}

type IoT struct {
	sdk *SDK
}

func (m *IoT) Devices() *devices.Devices {
	return devices.NewDevices(m.sdk.getConn(IoTDevicesServiceID))
}

func (m *IoT) Data() *data.Data {
	return data.NewData(m.sdk.getConn(IoTDataServiceID))
}

func (m *IoT) Broker() *broker.Broker {
	return broker.NewBroker(m.sdk.getConn(IoTBrokerServiceID))
}

func (m *IoT) BrokerData() *brokerdata.BrokerData {
	return brokerdata.NewBrokerData(m.sdk.getConn(IoTBrokerDataServiceID))
}
