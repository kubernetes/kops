load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "internal.go",
        "match.go",
        "tables.go",
    ],
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/text/language:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "gen_test.go",
        "internal_test.go",
        "match_test.go",
    ],
    library = ":go_default_library",
    deps = ["//vendor/golang.org/x/text/language:go_default_library"],
)
