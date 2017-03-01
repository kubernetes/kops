// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kafka

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/golang/glog"
	"github.com/optiopay/kafka"
	"github.com/optiopay/kafka/proto"
)

const (
	brokerClientID           = "kafka-sink"
	brokerDialTimeout        = 10 * time.Second
	brokerDialRetryLimit     = 1
	brokerDialRetryWait      = 0
	brokerAllowTopicCreation = true
	brokerLeaderRetryLimit   = 1
	brokerLeaderRetryWait    = 0
	metricsTopic             = "heapster-metrics"
	eventsTopic              = "heapster-events"
)

const (
	TimeSeriesTopic = "timeseriestopic"
	EventsTopic     = "eventstopic"
)

type KafkaClient interface {
	Name() string
	Stop()
	ProduceKafkaMessage(msgData interface{}) error
}

type kafkaSink struct {
	producer  kafka.DistributingProducer
	dataTopic string
}

func (sink *kafkaSink) ProduceKafkaMessage(msgData interface{}) error {
	start := time.Now()
	msgJson, err := json.Marshal(msgData)
	if err != nil {
		return fmt.Errorf("failed to transform the items to json : %s", err)
	}

	message := &proto.Message{Value: []byte(string(msgJson))}
	_, err = sink.producer.Distribute(sink.dataTopic, message)
	if err != nil {
		return fmt.Errorf("failed to produce message to %s: %s", sink.dataTopic, err)
	}
	end := time.Now()
	glog.V(4).Infof("Exported %d data to kafka in %s", len([]byte(string(msgJson))), end.Sub(start))
	return nil
}

func (sink *kafkaSink) Name() string {
	return "Apache Kafka Sink"
}

func (sink *kafkaSink) Stop() {
	// nothing needs to be done.
}

// setupProducer returns a producer of kafka server
func setupProducer(sinkBrokerHosts []string, topic string, brokerConf kafka.BrokerConf) (kafka.DistributingProducer, error) {
	glog.V(3).Infof("attempting to setup kafka sink")
	broker, err := kafka.Dial(sinkBrokerHosts, brokerConf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to kafka cluster: %s", err)
	}
	defer broker.Close()

	//create kafka producer
	conf := kafka.NewProducerConf()
	conf.RequiredAcks = proto.RequiredAcksLocal
	producer := broker.Producer(conf)

	// create RoundRobinProducer with the default producer.
	count, err := broker.PartitionCount(topic)
	if err != nil {
		count = 1
		glog.Warningf("Failed to get partition count of topic %q: %s", topic, err)
	}
	sinkProducer := kafka.NewRoundRobinProducer(producer, count)
	glog.V(3).Infof("kafka sink setup successfully")
	return sinkProducer, nil
}

func getTopic(opts map[string][]string, topicType string) (string, error) {
	var topic string
	switch topicType {
	case TimeSeriesTopic:
		topic = metricsTopic
	case EventsTopic:
		topic = eventsTopic
	default:
		return "", fmt.Errorf("Topic type '%s' is illegal.", topicType)
	}

	if len(opts[topicType]) > 0 {
		topic = opts[topicType][0]
	}

	return topic, nil
}

func NewKafkaClient(uri *url.URL, topicType string) (KafkaClient, error) {
	opts, err := url.ParseQuery(uri.RawQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to parser url's query string: %s", err)
	}
	glog.V(3).Infof("kafka sink option: %v", opts)

	topic, err := getTopic(opts, topicType)
	if err != nil {
		return nil, err
	}

	var kafkaBrokers []string
	if len(opts["brokers"]) < 1 {
		return nil, fmt.Errorf("There is no broker assigned for connecting kafka")
	}
	kafkaBrokers = append(kafkaBrokers, opts["brokers"]...)
	glog.V(2).Infof("initializing kafka sink with brokers - %v", kafkaBrokers)

	//structure the config of broker
	brokerConf := kafka.NewBrokerConf(brokerClientID)
	brokerConf.DialTimeout = brokerDialTimeout
	brokerConf.DialRetryLimit = brokerDialRetryLimit
	brokerConf.DialRetryWait = brokerDialRetryWait
	brokerConf.LeaderRetryLimit = brokerLeaderRetryLimit
	brokerConf.LeaderRetryWait = brokerLeaderRetryWait
	brokerConf.AllowTopicCreation = brokerAllowTopicCreation

	// set up producer of kafka server.
	sinkProducer, err := setupProducer(kafkaBrokers, topic, brokerConf)
	if err != nil {
		return nil, fmt.Errorf("Failed to setup Producer: - %v", err)
	}

	return &kafkaSink{
		producer:  sinkProducer,
		dataTopic: topic,
	}, nil
}
