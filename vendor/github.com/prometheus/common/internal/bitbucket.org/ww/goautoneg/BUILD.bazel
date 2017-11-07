load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["autoneg.go"],
    importpath = "github.com/prometheus/common/internal/bitbucket.org/ww/goautoneg",
    visibility = ["//vendor/github.com/prometheus/common:__subpackages__"],
)

go_test(
    name = "go_default_test",
    srcs = ["autoneg_test.go"],
    importpath = "github.com/prometheus/common/internal/bitbucket.org/ww/goautoneg",
    library = ":go_default_library",
)
