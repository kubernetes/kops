load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["yaml.go"],
    importpath = "github.com/go-openapi/loads/fmts",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/go-openapi/swag:go_default_library",
        "//vendor/gopkg.in/yaml.v2:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "fixture_test.go",
        "yaml_test.go",
    ],
    importpath = "github.com/go-openapi/loads/fmts",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/go-openapi/spec:go_default_library",
        "//vendor/github.com/go-openapi/swag:go_default_library",
        "//vendor/github.com/stretchr/testify/assert:go_default_library",
    ],
)
