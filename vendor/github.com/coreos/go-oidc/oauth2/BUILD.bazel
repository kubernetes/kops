load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "error.go",
        "oauth2.go",
    ],
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/coreos/go-oidc/http:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["oauth2_test.go"],
    library = ":go_default_library",
    deps = ["//vendor/github.com/coreos/go-oidc/http:go_default_library"],
)
