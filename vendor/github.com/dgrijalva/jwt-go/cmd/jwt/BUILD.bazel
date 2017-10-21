load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["app.go"],
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/dgrijalva/jwt-go:go_default_library"],
)

go_binary(
    name = "jwt",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
