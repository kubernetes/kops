load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "google.golang.org/grpc/benchmark/server",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/google.golang.org/grpc/benchmark:go_default_library",
        "//vendor/google.golang.org/grpc/grpclog:go_default_library",
    ],
)

go_binary(
    name = "server",
    importpath = "google.golang.org/grpc/benchmark/server",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
