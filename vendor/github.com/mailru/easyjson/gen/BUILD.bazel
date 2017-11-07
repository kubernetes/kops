load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "decoder.go",
        "encoder.go",
        "generator.go",
    ],
    importpath = "github.com/mailru/easyjson/gen",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/mailru/easyjson:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["generator_test.go"],
    importpath = "github.com/mailru/easyjson/gen",
    library = ":go_default_library",
)
