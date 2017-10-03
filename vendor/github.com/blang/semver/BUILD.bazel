load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "json.go",
        "range.go",
        "semver.go",
        "sort.go",
        "sql.go",
    ],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "json_test.go",
        "range_test.go",
        "semver_test.go",
        "sort_test.go",
        "sql_test.go",
    ],
    library = ":go_default_library",
)
