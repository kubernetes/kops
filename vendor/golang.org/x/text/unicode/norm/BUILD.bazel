load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "composition.go",
        "forminfo.go",
        "input.go",
        "iter.go",
        "normalize.go",
        "readwriter.go",
        "tables.go",
        "transform.go",
        "trie.go",
    ],
    importpath = "golang.org/x/text/unicode/norm",
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/text/transform:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "composition_test.go",
        "iter_test.go",
        "normalize_test.go",
        "readwriter_test.go",
        "transform_test.go",
        "ucd_test.go",
    ],
    importpath = "golang.org/x/text/unicode/norm",
    library = ":go_default_library",
    deps = [
        "//vendor/golang.org/x/text/internal/gen:go_default_library",
        "//vendor/golang.org/x/text/internal/testtext:go_default_library",
        "//vendor/golang.org/x/text/transform:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = [
        "example_iter_test.go",
        "example_test.go",
        "norm_test.go",
    ],
    importpath = "golang.org/x/text/unicode/norm_test",
    deps = [":go_default_library"],
)
