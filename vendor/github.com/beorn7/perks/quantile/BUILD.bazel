load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["stream.go"],
    importpath = "github.com/beorn7/perks/quantile",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "bench_test.go",
        "stream_test.go",
    ],
    importpath = "github.com/beorn7/perks/quantile",
    library = ":go_default_library",
)

go_test(
    name = "go_default_xtest",
    srcs = ["example_test.go"],
    importpath = "github.com/beorn7/perks/quantile_test",
    deps = [":go_default_library"],
)
