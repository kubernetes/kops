load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["proto3ext.pb.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/gogo/protobuf/proto:go_default_library",
        "//vendor/github.com/gogo/protobuf/protoc-gen-gogo/descriptor:go_default_library",
    ],
)

filegroup(
    name = "go_default_library_protos",
    srcs = ["proto3ext.proto"],
    visibility = ["//visibility:public"],
)
