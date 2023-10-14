# Copyright 2023 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: protobuf
protobuf:
	protoc --go_out=pb/ --go_opt=paths=source_relative --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	-I pb/ \
	pb/jaeger/api/v2/*.proto pb/jaeger/storage/v1/*.proto
	go run golang.org/x/tools/cmd/goimports@latest -w ./pb

.PHONY: wget
wget:
	mkdir -p pb/jaeger/api/v2/
	wget -O pb/jaeger/api/v2/model.proto https://raw.githubusercontent.com/jaegertracing/jaeger-idl/main/proto/api_v2/model.proto
	mkdir -p pb/jaeger/storage/v1/
	wget -O pb/jaeger/storage/v1/storage.proto https://raw.githubusercontent.com/jaegertracing/jaeger/main/plugin/storage/grpc/proto/storage.proto
	# Remove gogoproto
	sed -i '/gogoproto/d' pb/jaeger/api/v2/model.proto
	sed -i 's@go_package@go_package="k8s.io/kops/tools/otel/traceserver/pb/jaeger/api/v2"; //@g' pb/jaeger/api/v2/model.proto
	sed -i '/gogoproto/d' pb/jaeger/storage/v1/storage.proto
	sed -i 's@go_package@go_package="k8s.io/kops/tools/otel/traceserver/pb/jaeger/storage/v1"; //@g' pb/jaeger/storage/v1/storage.proto
	sed -i 's@import "model.proto"@import "jaeger/api/v2/model.proto"@g' pb/jaeger/storage/v1/storage.proto
	# Remove empty [ ] directives (which span lines)
	cat pb/jaeger/api/v2/model.proto | \
	 tr '\n' '~' | sed 's/\[\~[ ]*\]//g' | \
	 tr '~' '\n' > pb/jaeger/api/v2/model.proto.out
	mv pb/jaeger/api/v2/model.proto.out pb/jaeger/api/v2/model.proto
	cat pb/jaeger/storage/v1/storage.proto | \
	 tr '\n' '~' | sed 's/\[\~[ ]*\]//g' | \
	 tr '~' '\n' > pb/jaeger/storage/v1/storage.proto.out
	mv pb/jaeger/storage/v1/storage.proto.out pb/jaeger/storage/v1/storage.proto
