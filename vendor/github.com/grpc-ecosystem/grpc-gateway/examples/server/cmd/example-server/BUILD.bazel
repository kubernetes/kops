load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/examples/server/cmd/example-server",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/examples/server:go_default_library",
    ],
)

go_binary(
    name = "example-server",
    importpath = "github.com/grpc-ecosystem/grpc-gateway/examples/server/cmd/example-server",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
