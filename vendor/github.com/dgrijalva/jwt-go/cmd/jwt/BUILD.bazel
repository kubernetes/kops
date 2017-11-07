load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["app.go"],
    importpath = "github.com/dgrijalva/jwt-go/cmd/jwt",
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/dgrijalva/jwt-go:go_default_library"],
)

go_binary(
    name = "jwt",
    importpath = "github.com/dgrijalva/jwt-go/cmd/jwt",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
