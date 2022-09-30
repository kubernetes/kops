package ycsdk

import (
	logging "github.com/yandex-cloud/go-sdk/gen/logging"
	ingestion "github.com/yandex-cloud/go-sdk/gen/logingestion"
	reading "github.com/yandex-cloud/go-sdk/gen/logreading"
)

const (
	LoggingServiceID      Endpoint = "logging"
	LogReadingServiceID   Endpoint = "log-reading"
	LogIngestionServiceID Endpoint = "log-ingestion"
)

func (sdk *SDK) Logging() *logging.Logging {
	return logging.NewLogging(sdk.getConn(LoggingServiceID))
}

func (sdk *SDK) LogReading() *reading.LogReading {
	return reading.NewLogReading(sdk.getConn(LogReadingServiceID))
}

func (sdk *SDK) LogIngestion() *ingestion.LogIngestion {
	return ingestion.NewLogIngestion(sdk.getConn(LogIngestionServiceID))
}
