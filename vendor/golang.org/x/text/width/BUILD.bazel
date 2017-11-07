load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "kind_string.go",
        "tables.go",
        "transform.go",
        "trieval.go",
        "width.go",
    ],
    importpath = "golang.org/x/text/width",
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/text/transform:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "common_test.go",
        "runes_test.go",
        "tables_test.go",
        "transform_test.go",
    ],
    importpath = "golang.org/x/text/width",
    library = ":go_default_library",
    deps = [
        "//vendor/golang.org/x/text/internal/gen:go_default_library",
        "//vendor/golang.org/x/text/internal/testtext:go_default_library",
        "//vendor/golang.org/x/text/internal/ucd:go_default_library",
        "//vendor/golang.org/x/text/transform:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = ["example_test.go"],
    importpath = "golang.org/x/text/width_test",
    deps = [":go_default_library"],
)
