load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "EchoServiceApi.go",
        "ExamplepbSimpleMessage.go",
    ],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/examples/clients/echo",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/dghubble/sling:go_default_library"],
)
