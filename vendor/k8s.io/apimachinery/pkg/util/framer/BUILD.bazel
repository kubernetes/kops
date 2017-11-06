load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["framer.go"],
    importpath = "k8s.io/apimachinery/pkg/util/framer",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["framer_test.go"],
    importpath = "k8s.io/apimachinery/pkg/util/framer",
    library = ":go_default_library",
)
