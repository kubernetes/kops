load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "blake2b.go",
        "blake2b_generic.go",
        "blake2x.go",
        "register.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "blake2bAVX2_amd64.go",
            "blake2bAVX2_amd64.s",
            "blake2b_amd64.go",
            "blake2b_amd64.s",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "blake2bAVX2_amd64.go",
            "blake2bAVX2_amd64.s",
            "blake2b_amd64.go",
            "blake2b_amd64.s",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "blake2bAVX2_amd64.go",
            "blake2bAVX2_amd64.s",
            "blake2b_amd64.go",
            "blake2b_amd64.s",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/crypto/blake2b",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["blake2b_test.go"],
    importpath = "golang.org/x/crypto/blake2b",
    library = ":go_default_library",
)
