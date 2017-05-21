load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "generator.go",
        "main.go",
    ],
    visibility = ["//visibility:private"],
    deps = ["//upup/tools/generators/pkg/codegen:go_default_library"],
)

go_binary(
    name = "fitask",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
