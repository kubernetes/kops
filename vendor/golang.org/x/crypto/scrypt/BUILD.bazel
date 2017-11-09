load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["scrypt.go"],
    importpath = "golang.org/x/crypto/scrypt",
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/crypto/pbkdf2:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["scrypt_test.go"],
    importpath = "golang.org/x/crypto/scrypt",
    library = ":go_default_library",
)
