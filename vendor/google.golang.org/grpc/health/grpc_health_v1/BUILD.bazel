load("@io_bazel_rules_go//proto:def.bzl", "go_grpc_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["health.pb.go"],
    importpath = "google.golang.org/grpc/health/grpc_health_v1",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)

proto_library(
    name = "grpc_health_v1_proto",
    srcs = ["health.proto"],
    visibility = ["//visibility:public"],
)

go_grpc_library(
    name = "grpc_health_v1_go_proto",
    importpath = "google.golang.org/grpc/health/grpc_health_v1",
    proto = ":grpc_health_v1_proto",
    visibility = ["//visibility:public"],
)
