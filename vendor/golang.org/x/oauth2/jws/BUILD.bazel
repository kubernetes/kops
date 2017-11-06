load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["jws.go"],
    importpath = "golang.org/x/oauth2/jws",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["jws_test.go"],
    importpath = "golang.org/x/oauth2/jws",
    library = ":go_default_library",
)
