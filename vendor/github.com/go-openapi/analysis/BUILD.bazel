load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["analyzer.go"],
    importpath = "github.com/go-openapi/analysis",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/go-openapi/jsonpointer:go_default_library",
        "//vendor/github.com/go-openapi/spec:go_default_library",
        "//vendor/github.com/go-openapi/swag:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["analyzer_test.go"],
    importpath = "github.com/go-openapi/analysis",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/go-openapi/loads/fmts:go_default_library",
        "//vendor/github.com/go-openapi/spec:go_default_library",
        "//vendor/github.com/stretchr/testify/assert:go_default_library",
    ],
)
