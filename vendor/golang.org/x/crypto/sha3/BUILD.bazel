load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "hashes.go",
        "register.go",
        "sha3.go",
        "shake.go",
        "xor_generic.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "keccakf_amd64.go",
            "keccakf_amd64.s",
            "xor_unaligned.go",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "keccakf_amd64.go",
            "keccakf_amd64.s",
            "xor_unaligned.go",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "keccakf_amd64.go",
            "keccakf_amd64.s",
            "xor_unaligned.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/crypto/sha3",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["sha3_test.go"],
    data = glob(["testdata/**"]),
    importpath = "golang.org/x/crypto/sha3",
    library = ":go_default_library",
)
