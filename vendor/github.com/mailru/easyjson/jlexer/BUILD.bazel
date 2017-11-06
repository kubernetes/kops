load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "error.go",
        "lexer.go",
    ],
    importpath = "github.com/mailru/easyjson/jlexer",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["lexer_test.go"],
    importpath = "github.com/mailru/easyjson/jlexer",
    library = ":go_default_library",
)
