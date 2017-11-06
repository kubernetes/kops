load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "byte.go",
        "doc.go",
        "empty.go",
        "int.go",
        "int64.go",
        "string.go",
    ],
    importpath = "k8s.io/gengo/examples/set-gen/sets",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["set_test.go"],
    importpath = "k8s.io/gengo/examples/set-gen/sets",
    library = ":go_default_library",
)
