load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "pattern.go",
        "trie.go",
    ],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/utilities",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_xtest",
    srcs = ["trie_test.go"],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/utilities_test",
    deps = [":go_default_library"],
)
