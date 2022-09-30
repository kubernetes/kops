// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package ycsdk

import (
	"github.com/yandex-cloud/go-sdk/gen/mdb/clickhouse"
	"github.com/yandex-cloud/go-sdk/gen/mdb/elasticsearch"
	"github.com/yandex-cloud/go-sdk/gen/mdb/greenplum"
	"github.com/yandex-cloud/go-sdk/gen/mdb/kafka"
	"github.com/yandex-cloud/go-sdk/gen/mdb/mongodb"
	"github.com/yandex-cloud/go-sdk/gen/mdb/mysql"
	"github.com/yandex-cloud/go-sdk/gen/mdb/postgresql"
	"github.com/yandex-cloud/go-sdk/gen/mdb/redis"
	"github.com/yandex-cloud/go-sdk/gen/mdb/sqlserver"
)

const (
	MDBMongoDBServiceID    Endpoint = "managed-mongodb"
	MDBClickhouseServiceID Endpoint = "managed-clickhouse"
	MDBPostgreSQLServiceID Endpoint = "managed-postgresql"
	MDBRedisServiceID      Endpoint = "managed-redis"
	MDBMySQLServiceID      Endpoint = "managed-mysql"
	MDBKafkaServiceID      Endpoint = "managed-kafka"
	MDBSQLServerServiceID  Endpoint = "managed-sqlserver"
	MDBGreenplumServiceID  Endpoint = "managed-greenplum"
	MDBElasticSearchID     Endpoint = "managed-elasticsearch"
)

type MDB struct {
	sdk *SDK
}

func (m *MDB) PostgreSQL() *postgresql.PostgreSQL {
	return postgresql.NewPostgreSQL(m.sdk.getConn(MDBPostgreSQLServiceID))
}

func (m *MDB) MongoDB() *mongodb.MongoDB {
	return mongodb.NewMongoDB(m.sdk.getConn(MDBMongoDBServiceID))
}

func (m *MDB) Clickhouse() *clickhouse.Clickhouse {
	return clickhouse.NewClickhouse(m.sdk.getConn(MDBClickhouseServiceID))
}

func (m *MDB) Redis() *redis.Redis {
	return redis.NewRedis(m.sdk.getConn(MDBRedisServiceID))
}

func (m *MDB) Kafka() *kafka.Kafka {
	return kafka.NewKafka(m.sdk.getConn(MDBKafkaServiceID))
}

func (m *MDB) MySQL() *mysql.MySQL {
	return mysql.NewMySQL(m.sdk.getConn(MDBMySQLServiceID))
}

func (m *MDB) SQLServer() *sqlserver.SQLServer {
	return sqlserver.NewSQLServer(m.sdk.getConn(MDBSQLServerServiceID))
}

func (m *MDB) Greenplum() *greenplum.Greenplum {
	return greenplum.NewGreenplum(m.sdk.getConn(MDBGreenplumServiceID))
}

func (m *MDB) ElasticSearch() *elasticsearch.ElasticSearch {
	return elasticsearch.NewElasticSearch(m.sdk.getConn(MDBElasticSearchID))
}
