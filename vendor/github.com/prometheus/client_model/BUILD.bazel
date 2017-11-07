load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

proto_library(
    name = "io_prometheus_client_proto",
    srcs = ["metrics.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "io_prometheus_client_go_proto",
    importpath = "github.com/prometheus/client_model",
    proto = ":io_prometheus_client_proto",
    visibility = ["//visibility:public"],
)

go_library(
    name = "go_default_library",
    importpath = "github.com/prometheus/client_model",
    library = ":io_prometheus_client_go_proto",
    visibility = ["//visibility:public"],
)
