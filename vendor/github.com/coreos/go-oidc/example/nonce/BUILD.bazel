load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["app.go"],
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/coreos/go-oidc:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/oauth2:go_default_library",
    ],
)

go_binary(
    name = "nonce",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
