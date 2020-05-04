# Copyright 2019 The Kubernetes Authors.
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

FROM alpine:3.11 as ssl
RUN apk --no-cache add ca-certificates && \
    apk --no-cache upgrade

FROM scratch
LABEL Name=node-authorizer \
      Release=https://github.com/kubernetes/kops \
      Url=https://github.com/kubernetes/kops \
      Help=https://github.com/kubernetes/kops\issues

ADD bin/node-authorizer /node-authorizer
COPY --from=ssl /etc/ssl /etc/ssl

ENTRYPOINT ["/node-authorizer"]
