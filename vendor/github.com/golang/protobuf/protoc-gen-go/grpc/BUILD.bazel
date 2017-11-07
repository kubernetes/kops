load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["grpc.go"],
    importpath = "github.com/golang/protobuf/protoc-gen-go/grpc",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/protoc-gen-go/descriptor:go_default_library",
        "//vendor/github.com/golang/protobuf/protoc-gen-go/generator:go_default_library",
    ],
)
