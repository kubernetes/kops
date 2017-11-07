load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "base64.go",
        "bcrypt.go",
    ],
    importpath = "golang.org/x/crypto/bcrypt",
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/crypto/blowfish:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["bcrypt_test.go"],
    importpath = "golang.org/x/crypto/bcrypt",
    library = ":go_default_library",
)
