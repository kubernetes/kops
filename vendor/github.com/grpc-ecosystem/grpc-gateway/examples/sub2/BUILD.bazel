load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["message.pb.go"],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/examples/sub2",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/protobuf/proto:go_default_library"],
)

proto_library(
    name = "sub2_proto",
    srcs = ["message.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "sub2_go_proto",
    importpath = "github.com/grpc-ecosystem/grpc-gateway/examples/sub2",
    proto = ":sub2_proto",
    visibility = ["//visibility:public"],
)
