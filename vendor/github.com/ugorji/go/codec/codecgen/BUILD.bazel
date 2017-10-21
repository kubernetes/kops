load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "gen.go",
        "z.go",
    ],
    visibility = ["//visibility:private"],
)

go_binary(
    name = "codecgen",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
