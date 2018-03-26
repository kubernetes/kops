load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["ed25519.go"],
    importpath = "golang.org/x/crypto/ed25519",
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/crypto/ed25519/internal/edwards25519:go_default_library"],
)
