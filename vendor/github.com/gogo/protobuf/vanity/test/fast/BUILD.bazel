load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "gogovanity.pb.go",
        "proto3.pb.go",
        "vanity.pb.go",
    ],
    importpath = "github.com/gogo/protobuf/vanity/test/fast",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/gogo/protobuf/gogoproto:go_default_library",
        "//vendor/github.com/gogo/protobuf/proto:go_default_library",
    ],
)
