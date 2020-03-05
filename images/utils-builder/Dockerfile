# Copyright 2017 The Kubernetes Authors.
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

# Sadly we can't update to 0.3 (stretch), because that version FTBFS statically
FROM k8s.gcr.io/debian-base-amd64:0.2

RUN echo "deb http://deb.debian.org/debian jessie main\n \
deb http://archive.debian.org/debian/ jessie main\n \
deb-src http://archive.debian.org/debian/ jessie main\n \
deb http://security.debian.org jessie/updates main\n \
deb-src http://security.debian.org jessie/updates main\n \
deb-src http://ftp.us.debian.org/debian/ jessie main" > /etc/apt/sources.list

RUN apt-get update && apt-get --no-install-recommends install --yes dpkg-dev bash \
  && apt-get build-dep --yes socat conntrack \
  && apt-get clean

RUN mkdir /socat

# Note that this approach does _not_ include libssl, but we don't need it for kubernetes anyway
RUN cd /socat; \
    CFLAGS=-static LDFLAGS=-static CPPFLAGS=-static CFLAGS_APPEND=-static \
    LDFLAGS_APPEND=-static CPPFLAGS_APPEND=-static \
    apt-get source --build socat

RUN mkdir /conntrack

# Note that this approach does _not_ include libssl, but we don't need it for kubernetes anyway
RUN cd /conntrack; \
    CFLAGS=-static LDFLAGS=-static CPPFLAGS=-static CFLAGS_APPEND=-static \
    LDFLAGS_APPEND=-static CPPFLAGS_APPEND=-static \
    apt-get source --build conntrack

COPY extract.sh /extract.sh
