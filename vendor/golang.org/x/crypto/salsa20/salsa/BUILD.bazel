load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "hsalsa20.go",
        "salsa208.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "salsa2020_amd64.s",
            "salsa20_amd64.go",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "salsa2020_amd64.s",
            "salsa20_amd64.go",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "salsa2020_amd64.s",
            "salsa20_amd64.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/crypto/salsa20/salsa",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["salsa_test.go"],
    importpath = "golang.org/x/crypto/salsa20/salsa",
    library = ":go_default_library",
)
