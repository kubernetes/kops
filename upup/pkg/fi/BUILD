load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "assetstore.go",
        "ca.go",
        "changes.go",
        "cloud.go",
        "compare_with_id.go",
        "context.go",
        "default_methods.go",
        "deletions.go",
        "dryrun_target.go",
        "errors.go",
        "executor.go",
        "files.go",
        "has_address.go",
        "http.go",
        "named.go",
        "resources.go",
        "secrets.go",
        "target.go",
        "task.go",
        "timestamp.go",
        "topological_sort.go",
        "users.go",
        "values.go",
        "vfs_castore.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//forks/pkg/dnsprovider:go_default_library",
        "//pkg/diff:go_default_library",
        "//upup/pkg/fi/utils:go_default_library",
        "//util/pkg/hashing:go_default_library",
        "//util/pkg/vfs:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/api/validation:go_default_library",
        "@io_k8s_apimachinery//pkg/util/validation/field:go_default_library",
        "@org_golang_x_crypto//ssh:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["vfs_castore_test.go"],
    library = ":go_default_library",
)
