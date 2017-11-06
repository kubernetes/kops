load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["common.pb.go"],
    importpath = "google.golang.org/genproto/googleapis/cloud/support/common",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/timestamp:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/annotations:go_default_library",
    ],
)
