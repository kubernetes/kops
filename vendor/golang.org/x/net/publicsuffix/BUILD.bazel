load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "list.go",
        "table.go",
    ],
    importpath = "golang.org/x/net/publicsuffix",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "list_test.go",
        "table_test.go",
    ],
    importpath = "golang.org/x/net/publicsuffix",
    library = ":go_default_library",
)
