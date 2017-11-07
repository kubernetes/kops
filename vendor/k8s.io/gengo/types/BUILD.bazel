load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "comments.go",
        "doc.go",
        "flatten.go",
        "types.go",
    ],
    importpath = "k8s.io/gengo/types",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "comments_test.go",
        "flatten_test.go",
        "types_test.go",
    ],
    importpath = "k8s.io/gengo/types",
    library = ":go_default_library",
)
