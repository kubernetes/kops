load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["memcache.go"],
    importpath = "k8s.io/client-go/discovery/cached",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/emicklei/go-restful-swagger12:go_default_library",
        "//vendor/github.com/googleapis/gnostic/OpenAPIv2:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/schema:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/version:go_default_library",
        "//vendor/k8s.io/client-go/discovery:go_default_library",
        "//vendor/k8s.io/client-go/rest:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["memcache_test.go"],
    importpath = "k8s.io/client-go/discovery/cached",
    library = ":go_default_library",
    deps = [
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/client-go/discovery/fake:go_default_library",
    ],
)
