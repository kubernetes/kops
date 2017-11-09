load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "gen.go",
        "z.go",
    ],
    importpath = "github.com/ugorji/go/codec/codecgen",
    visibility = ["//visibility:private"],
)

go_binary(
    name = "codecgen",
    importpath = "github.com/ugorji/go/codec/codecgen",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
