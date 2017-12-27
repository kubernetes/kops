load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "helpers.go",
        "raw.go",
    ],
    importpath = "github.com/mailru/easyjson",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/mailru/easyjson/jlexer:go_default_library",
        "//vendor/github.com/mailru/easyjson/jwriter:go_default_library",
    ],
)
