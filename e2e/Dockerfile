# Copyright 2016 The Kubernetes Authors.
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

FROM debian:jessie

RUN apt-get update && apt-get install --yes curl python-pip openssh-client
RUN pip install awscli

RUN curl https://sdk.cloud.google.com | bash

RUN curl https://storage.googleapis.com/golang/go1.6.3.linux-amd64.tar.gz | tar -C /usr/local -xzf -

ENV PATH /root/google-cloud-sdk/bin:/usr/local/go/bin:$PATH

ADD runtests.sh /

ADD conf /conf
