load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["encoding.go"],
    importpath = "golang.org/x/text/encoding",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/golang.org/x/text/encoding/internal/identifier:go_default_library",
        "//vendor/golang.org/x/text/transform:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = [
        "encoding_test.go",
        "example_test.go",
    ],
    data = glob(["testdata/**"]),
    importpath = "golang.org/x/text/encoding_test",
    deps = [
        ":go_default_library",
        "//vendor/golang.org/x/text/encoding/charmap:go_default_library",
        "//vendor/golang.org/x/text/encoding/unicode:go_default_library",
        "//vendor/golang.org/x/text/transform:go_default_library",
    ],
)
