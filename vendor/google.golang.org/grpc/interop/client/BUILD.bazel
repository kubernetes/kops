load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["client.go"],
    importpath = "google.golang.org/grpc/interop/client",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/credentials:go_default_library",
        "//vendor/google.golang.org/grpc/credentials/oauth:go_default_library",
        "//vendor/google.golang.org/grpc/grpclog:go_default_library",
        "//vendor/google.golang.org/grpc/interop:go_default_library",
        "//vendor/google.golang.org/grpc/interop/grpc_testing:go_default_library",
    ],
)

go_binary(
    name = "client",
    importpath = "google.golang.org/grpc/interop/client",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
