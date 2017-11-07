load("@io_bazel_rules_go//proto:def.bzl", "go_grpc_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["test.pb.go"],
    importpath = "github.com/grpc-ecosystem/go-grpc-prometheus/examples/testproto",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)

proto_library(
    name = "mwitkow_testproto_proto",
    srcs = ["test.proto"],
    visibility = ["//visibility:public"],
)

go_grpc_library(
    name = "mwitkow_testproto_go_proto",
    importpath = "github.com/grpc-ecosystem/go-grpc-prometheus/examples/testproto",
    proto = ":mwitkow_testproto_proto",
    visibility = ["//visibility:public"],
)
