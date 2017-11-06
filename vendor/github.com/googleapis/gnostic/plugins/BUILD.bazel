load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["plugin.pb.go"],
    importpath = "github.com/googleapis/gnostic/plugins",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/protobuf/proto:go_default_library"],
)

proto_library(
    name = "openapi_plugin_v1_proto",
    srcs = ["plugin.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "openapi_plugin_v1_go_proto",
    importpath = "github.com/googleapis/gnostic/plugins",
    proto = ":openapi_plugin_v1_proto",
    visibility = ["//visibility:public"],
)
