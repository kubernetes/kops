load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["ocsp.go"],
    importpath = "golang.org/x/crypto/ocsp",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["ocsp_test.go"],
    importpath = "golang.org/x/crypto/ocsp",
    library = ":go_default_library",
)
