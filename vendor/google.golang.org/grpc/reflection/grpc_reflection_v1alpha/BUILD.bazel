load("@io_bazel_rules_go//proto:def.bzl", "go_grpc_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["reflection.pb.go"],
    importpath = "google.golang.org/grpc/reflection/grpc_reflection_v1alpha",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)

proto_library(
    name = "grpc_reflection_v1alpha_proto",
    srcs = ["reflection.proto"],
    visibility = ["//visibility:public"],
)

go_grpc_library(
    name = "grpc_reflection_v1alpha_go_proto",
    importpath = "google.golang.org/grpc/reflection/grpc_reflection_v1alpha",
    proto = ":grpc_reflection_v1alpha_proto",
    visibility = ["//visibility:public"],
)
