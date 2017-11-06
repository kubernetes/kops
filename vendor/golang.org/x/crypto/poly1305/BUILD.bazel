load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "poly1305.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "sum_amd64.go",
            "sum_amd64.s",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "sum_amd64.go",
            "sum_amd64.s",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "sum_amd64.go",
            "sum_amd64.s",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/crypto/poly1305",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["poly1305_test.go"],
    importpath = "golang.org/x/crypto/poly1305",
    library = ":go_default_library",
)
