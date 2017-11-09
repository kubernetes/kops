load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["timestamp.pb.go"],
    importpath = "github.com/golang/protobuf/ptypes/timestamp",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/protobuf/proto:go_default_library"],
)

proto_library(
    name = "timestamp_proto",
    srcs = ["timestamp.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "timestamp_go_proto",
    importpath = "github.com/golang/protobuf/ptypes/timestamp",
    proto = ":timestamp_proto",
    visibility = ["//visibility:public"],
)
