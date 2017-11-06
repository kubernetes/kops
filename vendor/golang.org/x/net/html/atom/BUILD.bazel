load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "atom.go",
        "table.go",
    ],
    importpath = "golang.org/x/net/html/atom",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "atom_test.go",
        "table_test.go",
    ],
    importpath = "golang.org/x/net/html/atom",
    library = ":go_default_library",
)
