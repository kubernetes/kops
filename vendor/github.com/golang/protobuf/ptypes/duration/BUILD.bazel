load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["duration.pb.go"],
    importpath = "github.com/golang/protobuf/ptypes/duration",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/protobuf/proto:go_default_library"],
)

proto_library(
    name = "duration_proto",
    srcs = ["duration.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "duration_go_proto",
    importpath = "github.com/golang/protobuf/ptypes/duration",
    proto = ":duration_proto",
    visibility = ["//visibility:public"],
)
