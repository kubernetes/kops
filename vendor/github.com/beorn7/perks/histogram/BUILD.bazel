load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["histogram.go"],
    importpath = "github.com/beorn7/perks/histogram",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "bench_test.go",
        "histogram_test.go",
    ],
    importpath = "github.com/beorn7/perks/histogram",
    library = ":go_default_library",
)
