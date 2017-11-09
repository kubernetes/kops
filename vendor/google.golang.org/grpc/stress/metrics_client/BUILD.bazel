load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "google.golang.org/grpc/stress/metrics_client",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/grpclog:go_default_library",
        "//vendor/google.golang.org/grpc/stress/grpc_testing:go_default_library",
    ],
)

go_binary(
    name = "metrics_client",
    importpath = "google.golang.org/grpc/stress/metrics_client",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
