load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "cache_simple.go",
        "cache_striped.go",
        "cached_token_authenticator.go",
    ],
    importpath = "k8s.io/apiserver/pkg/authentication/token/cache",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/k8s.io/apimachinery/pkg/util/cache:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/clock:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/authentication/authenticator:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/authentication/user:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "cache_test.go",
        "cached_token_authenticator_test.go",
    ],
    importpath = "k8s.io/apiserver/pkg/authentication/token/cache",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/pborman/uuid:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/clock:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/authentication/authenticator:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/authentication/user:go_default_library",
    ],
)
