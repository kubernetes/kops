load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "fake_handler.go",
        "tmpdir.go",
    ],
    importpath = "k8s.io/client-go/util/testing",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["fake_handler_test.go"],
    importpath = "k8s.io/client-go/util/testing",
    library = ":go_default_library",
)
