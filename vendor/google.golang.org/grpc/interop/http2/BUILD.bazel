load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["negative_http2_client.go"],
    importpath = "google.golang.org/grpc/interop/http2",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "//vendor/google.golang.org/grpc/grpclog:go_default_library",
        "//vendor/google.golang.org/grpc/interop:go_default_library",
        "//vendor/google.golang.org/grpc/interop/grpc_testing:go_default_library",
    ],
)

go_binary(
    name = "http2",
    importpath = "google.golang.org/grpc/interop/http2",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
