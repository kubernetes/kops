load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["logger.go"],
    importmap = "vendor/google.golang.org/grpc/grpclog",
    importpath = "google.golang.org/grpc/grpclog",
    visibility = ["//visibility:public"],
)
