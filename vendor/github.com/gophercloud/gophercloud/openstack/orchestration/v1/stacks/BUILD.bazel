load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "environment.go",
        "errors.go",
        "fixtures.go",
        "requests.go",
        "results.go",
        "template.go",
        "urls.go",
        "utils.go",
    ],
    importpath = "github.com/gophercloud/gophercloud/openstack/orchestration/v1/stacks",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/gophercloud/gophercloud:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/pagination:go_default_library",
        "//vendor/gopkg.in/yaml.v2:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "environment_test.go",
        "template_test.go",
        "utils_test.go",
    ],
    importpath = "github.com/gophercloud/gophercloud/openstack/orchestration/v1/stacks",
    library = ":go_default_library",
    deps = ["//vendor/github.com/gophercloud/gophercloud/testhelper:go_default_library"],
)
