load("@io_bazel_rules_go//go:def.bzl", "go_test")

go_test(
    name = "go_default_xtest",
    srcs = ["hw_test.go"],
    importpath = "google.golang.org/grpc/examples/helloworld/mock_test",
    deps = [
        "//vendor/github.com/golang/mock/gomock:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc/examples/helloworld/helloworld:go_default_library",
        "//vendor/google.golang.org/grpc/examples/helloworld/mock/mock_helloworld:go_default_library",
    ],
)
