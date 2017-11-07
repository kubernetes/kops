load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "man_docs.go",
        "md_docs.go",
        "util.go",
    ],
    importpath = "github.com/spf13/cobra/doc",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/cpuguy83/go-md2man/md2man:go_default_library",
        "//vendor/github.com/spf13/cobra:go_default_library",
        "//vendor/github.com/spf13/pflag:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "cmd_test.go",
        "man_docs_test.go",
        "md_docs_test.go",
    ],
    importpath = "github.com/spf13/cobra/doc",
    library = ":go_default_library",
    deps = ["//vendor/github.com/spf13/cobra:go_default_library"],
)

go_test(
    name = "go_default_xtest",
    srcs = ["man_examples_test.go"],
    importpath = "github.com/spf13/cobra/doc_test",
    deps = [
        ":go_default_library",
        "//vendor/github.com/spf13/cobra:go_default_library",
    ],
)
