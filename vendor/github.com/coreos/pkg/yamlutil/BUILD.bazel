load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["yaml.go"],
    importpath = "github.com/coreos/pkg/yamlutil",
    visibility = ["//visibility:public"],
    deps = ["//vendor/gopkg.in/yaml.v1:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["yaml_test.go"],
    importpath = "github.com/coreos/pkg/yamlutil",
    library = ":go_default_library",
)
