load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "dlp.pb.go",
        "storage.pb.go",
    ],
    importpath = "google.golang.org/genproto/googleapis/privacy/dlp/v2beta1",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/empty:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/timestamp:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/annotations:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/longrunning:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/type/date:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/type/timeofday:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)
