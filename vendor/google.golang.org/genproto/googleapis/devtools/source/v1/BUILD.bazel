load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["source_context.pb.go"],
    importpath = "google.golang.org/genproto/googleapis/devtools/source/v1",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/annotations:go_default_library",
    ],
)
