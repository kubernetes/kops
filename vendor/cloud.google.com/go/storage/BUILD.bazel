load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "acl.go",
        "bucket.go",
        "reader.go",
        "storage.go",
        "writer.go",
    ],
    importpath = "cloud.google.com/go/storage",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/googleapi:go_default_library",
        "//vendor/google.golang.org/api/iterator:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/api/storage/v1:go_default_library",
        "//vendor/google.golang.org/api/transport:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "integration_test.go",
        "storage_test.go",
        "writer_test.go",
    ],
    data = glob(["testdata/**"]),
    importpath = "cloud.google.com/go/storage",
    library = ":go_default_library",
    deps = [
        "//vendor/cloud.google.com/go/internal/testutil:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/googleapi:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/api/storage/v1:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = ["example_test.go"],
    data = glob(["testdata/**"]),
    importpath = "cloud.google.com/go/storage_test",
    deps = [
        ":go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
