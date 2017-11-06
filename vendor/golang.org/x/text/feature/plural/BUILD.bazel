load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "common.go",
        "plural.go",
        "tables.go",
    ],
    importpath = "golang.org/x/text/feature/plural",
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/text/language:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "data_test.go",
        "plural_test.go",
    ],
    importpath = "golang.org/x/text/feature/plural",
    library = ":go_default_library",
    deps = ["//vendor/golang.org/x/text/language:go_default_library"],
)
