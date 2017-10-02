load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["proto3.pb.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/gogo/protobuf/proto:go_default_library",
        "//vendor/github.com/gogo/protobuf/proto/testdata:go_default_library",
        "//vendor/github.com/gogo/protobuf/types:go_default_library",
    ],
)

filegroup(
    name = "go_default_library_protos",
    srcs = ["proto3.proto"],
    visibility = ["//visibility:public"],
)
