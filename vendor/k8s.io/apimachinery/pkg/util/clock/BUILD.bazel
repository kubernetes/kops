load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["clock.go"],
    importpath = "k8s.io/apimachinery/pkg/util/clock",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["clock_test.go"],
    importpath = "k8s.io/apimachinery/pkg/util/clock",
    library = ":go_default_library",
)
