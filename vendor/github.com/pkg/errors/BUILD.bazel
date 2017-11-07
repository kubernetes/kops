load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "errors.go",
        "stack.go",
    ],
    importpath = "github.com/pkg/errors",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "bench_test.go",
        "errors_test.go",
        "format_test.go",
        "stack_test.go",
    ],
    importpath = "github.com/pkg/errors",
    library = ":go_default_library",
)

go_test(
    name = "go_default_xtest",
    srcs = ["example_test.go"],
    importpath = "github.com/pkg/errors_test",
    deps = [":go_default_library"],
)
