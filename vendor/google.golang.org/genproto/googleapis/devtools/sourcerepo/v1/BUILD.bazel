load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["sourcerepo.pb.go"],
    importpath = "google.golang.org/genproto/googleapis/devtools/sourcerepo/v1",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/empty:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/annotations:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/serviceconfig:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/iam/v1:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)
