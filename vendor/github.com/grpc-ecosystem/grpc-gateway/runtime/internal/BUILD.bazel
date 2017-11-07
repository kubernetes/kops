load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["stream_chunk.pb.go"],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/runtime/internal",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/protobuf/proto:go_default_library"],
)

proto_library(
    name = "internal_proto",
    srcs = ["stream_chunk.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "internal_go_proto",
    importpath = "github.com/grpc-ecosystem/grpc-gateway/runtime/internal",
    proto = ":internal_proto",
    visibility = ["//visibility:public"],
)
