load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "datastore.go",
        "doc.go",
        "errors.go",
        "key.go",
        "load.go",
        "prop.go",
        "query.go",
        "save.go",
        "time.go",
        "transaction.go",
    ],
    importpath = "cloud.google.com/go/datastore",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/timestamp:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/wrappers:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/api/transport:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/datastore/v1:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/type/latlng:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "//vendor/google.golang.org/grpc/metadata:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "datastore_test.go",
        "integration_test.go",
        "key_test.go",
        "load_test.go",
        "query_test.go",
        "save_test.go",
        "time_test.go",
    ],
    data = glob(["testdata/**"]),
    importpath = "cloud.google.com/go/datastore",
    library = ":go_default_library",
    deps = [
        "//vendor/cloud.google.com/go/internal/testutil:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/datastore/v1:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = [
        "example_test.go",
        "examples_test.go",
    ],
    data = glob(["testdata/**"]),
    importpath = "cloud.google.com/go/datastore_test",
    deps = [
        ":go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
