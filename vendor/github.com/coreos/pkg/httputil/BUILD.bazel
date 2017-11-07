load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "cookie.go",
        "json.go",
    ],
    importpath = "github.com/coreos/pkg/httputil",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "cookie_test.go",
        "json_test.go",
    ],
    importpath = "github.com/coreos/pkg/httputil",
    library = ":go_default_library",
)
