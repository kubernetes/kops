load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["monitored_resource.pb.go"],
    importpath = "google.golang.org/genproto/googleapis/api/monitoredres",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/label:go_default_library",
    ],
)
