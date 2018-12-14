# Buildifier

buildifier is a tool for formatting bazel BUILD files with a standard convention.

linux-x86_64 | ubuntu_15.10-x86_64 | darwin-x86_64
:---: | :---: | :---:
[![Build Status](http://ci.bazel.io/buildStatus/icon?job=buildifier/BAZEL_VERSION=latest,PLATFORM_NAME=linux-x86_64)](http://ci.bazel.io/job/buildifier/BAZEL_VERSION=latest,PLATFORM_NAME=linux-x86_64) | [![Build Status](http://ci.bazel.io/buildStatus/icon?job=buildifier/BAZEL_VERSION=latest,PLATFORM_NAME=ubuntu_15.10-x86_64)](http://ci.bazel.io/job/buildifier/BAZEL_VERSION=latest,PLATFORM_NAME=ubuntu_15.10-x86_64) | [![Build Status](http://ci.bazel.io/buildStatus/icon?job=buildifier/BAZEL_VERSION=latest,PLATFORM_NAME=darwin-x86_64)](http://ci.bazel.io/job/buildifier/BAZEL_VERSION=latest,PLATFORM_NAME=darwin-x86_64)

## Setup

Build the tool:
* Checkout the repo and then either via `go install` or `bazel build //buildifier`
* If you already have 'go' installed, then build a binary via:

`go get github.com/bazelbuild/buildtools/buildifier`

## Usage

Use buildifier to create standardized formatting for BUILD files in the
same way that clang-format is used for source files.

`$ buildifier -showlog -mode=check $(find . -iname BUILD -type f)`
