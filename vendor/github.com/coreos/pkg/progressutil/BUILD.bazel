load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "iocopy.go",
        "progressbar.go",
    ],
    importpath = "github.com/coreos/pkg/progressutil",
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/crypto/ssh/terminal:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["progressbar_test.go"],
    importpath = "github.com/coreos/pkg/progressutil",
    library = ":go_default_library",
)
