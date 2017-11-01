load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["empty.pb.go"],
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/protobuf/proto:go_default_library"],
)

filegroup(
    name = "go_default_library_protos",
    srcs = ["empty.proto"],
    visibility = ["//visibility:public"],
)
