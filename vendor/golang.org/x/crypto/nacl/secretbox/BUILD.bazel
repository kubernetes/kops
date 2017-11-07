load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["secretbox.go"],
    importpath = "golang.org/x/crypto/nacl/secretbox",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/golang.org/x/crypto/poly1305:go_default_library",
        "//vendor/golang.org/x/crypto/salsa20/salsa:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["secretbox_test.go"],
    importpath = "golang.org/x/crypto/nacl/secretbox",
    library = ":go_default_library",
)

go_test(
    name = "go_default_xtest",
    srcs = ["example_test.go"],
    importpath = "golang.org/x/crypto/nacl/secretbox_test",
    deps = [":go_default_library"],
)
