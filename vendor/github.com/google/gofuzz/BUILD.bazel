load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "fuzz.go",
    ],
    importpath = "github.com/google/gofuzz",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["fuzz_test.go"],
    importpath = "github.com/google/gofuzz",
    library = ":go_default_library",
)

go_test(
    name = "go_default_xtest",
    srcs = ["example_test.go"],
    importpath = "github.com/google/gofuzz_test",
    deps = [":go_default_library"],
)
