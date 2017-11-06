load("@io_bazel_rules_go//proto:def.bzl", "go_grpc_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["route_guide.pb.go"],
    importpath = "google.golang.org/grpc/examples/route_guide/routeguide",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)

proto_library(
    name = "routeguide_proto",
    srcs = ["route_guide.proto"],
    visibility = ["//visibility:public"],
)

go_grpc_library(
    name = "routeguide_go_proto",
    importpath = "google.golang.org/grpc/examples/route_guide/routeguide",
    proto = ":routeguide_proto",
    visibility = ["//visibility:public"],
)
