load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["grpc.go"],
    importpath = "github.com/gogo/protobuf/protoc-gen-gogo/grpc",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/gogo/protobuf/protoc-gen-gogo/descriptor:go_default_library",
        "//vendor/github.com/gogo/protobuf/protoc-gen-gogo/generator:go_default_library",
    ],
)
