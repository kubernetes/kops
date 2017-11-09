load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["helpers.go"],
    importpath = "github.com/dgrijalva/jwt-go/test",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/dgrijalva/jwt-go:go_default_library"],
)
