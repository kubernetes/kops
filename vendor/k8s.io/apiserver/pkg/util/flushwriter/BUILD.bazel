load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "writer.go",
    ],
    importpath = "k8s.io/apiserver/pkg/util/flushwriter",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["writer_test.go"],
    importpath = "k8s.io/apiserver/pkg/util/flushwriter",
    library = ":go_default_library",
)
