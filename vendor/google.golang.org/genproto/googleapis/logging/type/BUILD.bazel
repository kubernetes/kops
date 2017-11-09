load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "http_request.pb.go",
        "log_severity.pb.go",
    ],
    importpath = "google.golang.org/genproto/googleapis/logging/type",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/duration:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/annotations:go_default_library",
    ],
)
