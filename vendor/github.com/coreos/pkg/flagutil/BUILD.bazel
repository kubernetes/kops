load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "env.go",
        "env_file.go",
        "types.go",
    ],
    importpath = "github.com/coreos/pkg/flagutil",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "env_test.go",
        "file_env_test.go",
        "types_test.go",
    ],
    importpath = "github.com/coreos/pkg/flagutil",
    library = ":go_default_library",
)
