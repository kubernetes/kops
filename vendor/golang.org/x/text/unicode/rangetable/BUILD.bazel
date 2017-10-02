load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "merge.go",
        "rangetable.go",
        "tables.go",
    ],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "merge_test.go",
        "rangetable_test.go",
    ],
    library = ":go_default_library",
)
