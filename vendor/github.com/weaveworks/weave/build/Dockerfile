FROM golang:1.7.4

# Support Raspberry Pi 2 and newer
ENV GOARM 7

# The names of the architectures that we're fetching cross-compilers for from the Ubuntu packages
ENV DEB_CROSSPLATFORMS \
	armhf \
	arm64

# The go platforms we are supporting to cross-compile
ENV GO_CROSSPLATFORMS \
	linux/arm \
	linux/arm64

# The names of the gcc cross-compilers we have installed
ENV GCC_CROSSCOMPILERS \
	arm-linux-gnueabihf \
	aarch64-linux-gnu

# Temporarily add the Ubuntu repositories, because we're going to install gcc cross-compilers from there
# Install the build-essential and crossbuild-essential-ARCH packages
RUN echo "deb http://archive.ubuntu.com/ubuntu xenial main universe" > /etc/apt/sources.list.d/cgocrosscompiling.list \
  && apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 40976EAF437D05B5 3B4FE6ACC0B21F32 \
  && apt-get update \
  && apt-get install -y build-essential \
  && for platform in ${DEB_CROSSPLATFORMS}; do apt-get install -y crossbuild-essential-${platform}; done \
  && rm /etc/apt/sources.list.d/cgocrosscompiling.list \
  && apt-get update \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

# Install some required packages
# libpcap is required because we're linking against its C libraries from the prog/weaver binary
# flex and bison are required packages for compiling libpcap manually later
RUN apt-get update \
	&& apt-get install -y \
		libpcap-dev \
		python-requests \
		time \
		file \
		flex \
		bison

# Install common Go tools
RUN go get \
	github.com/golang/lint/golint \
	github.com/fzipp/gocyclo \
	gopkg.in/mvdan/sh.v1/cmd/shfmt \
	github.com/fatih/hclfmt \
	github.com/client9/misspell/cmd/misspell

# First clean, because we're gonna rebuild the std both with -race enabled and disabled
# Running go install std twice is intentional and required, as content is written in two different directories
# (/usr/local/go/pkg/linux_amd64 vs /usr/local/go/pkg/linux_amd64_race)
# The other architectures do not have support for -race
RUN go clean -i net \
	&& go install -tags netgo std \
	&& go install -race -tags netgo std

# Prebuild std for the other architectures
RUN for platform in ${GO_CROSSPLATFORMS}; do \
		GOOS=${platform%/*} GOARCH=${platform##*/} \
			go install -tags netgo std; done

# Allow full write access to the Go folders for anyone
RUN chmod -R a+w /usr/local/go

# The libpcap version from Debian packages is 1.6.2, matching that version here although newer versions of libpcap have been released
# We have to cross-compile the libpcap library for the new architectures, that's what we're doing here
ENV LIBPCAP_CROSS_VERSION=1.6.2
RUN curl -sSL http://www.tcpdump.org/release/libpcap-${LIBPCAP_CROSS_VERSION}.tar.gz | tar -xz \
	&& cd libpcap-${LIBPCAP_CROSS_VERSION} \
	&& for crosscompiler in ${GCC_CROSSCOMPILERS}; do \
		CC=${crosscompiler}-gcc ac_cv_linux_vers=2 ./configure --host=${crosscompiler} --with-pcap=linux \
		&& make \
		&& export LIB_DIR="/usr/local/lib/${crosscompiler}-gcc" VER="$(cat ./VERSION)" MAJOR_VER="$(sed 's/\([0-9][0-9]*\)\..*/\1/' ./VERSION)" \
		&& mkdir -p ${LIB_DIR} \
		&& cp -f libpcap.a libpcap.so.${VER} ${LIB_DIR} \
		&& ln -sf libpcap.so.${VER} /usr/local/lib/libpcap.so.${MAJOR_VER} \
		&& ln -sf libpcap.so.${MAJOR_VER} /usr/local/lib/libpcap.so \
		&& make clean; done

COPY build.sh /
ENTRYPOINT ["sh", "/build.sh"]
