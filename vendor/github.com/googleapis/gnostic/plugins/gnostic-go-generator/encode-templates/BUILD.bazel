load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/googleapis/gnostic/plugins/gnostic-go-generator/encode-templates",
    visibility = ["//visibility:private"],
)

go_binary(
    name = "encode-templates",
    importpath = "github.com/googleapis/gnostic/plugins/gnostic-go-generator/encode-templates",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
