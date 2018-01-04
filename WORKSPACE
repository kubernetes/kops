#=============================================================================
# Go rules

http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.8.1/rules_go-0.8.1.tar.gz",
    sha256 = "90bb270d0a92ed5c83558b2797346917c46547f6f7103e648941ecdb6b9d0e72",
)

load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains")
go_rules_dependencies()
go_register_toolchains()

#=============================================================================
# Docker rules

git_repository(
    name = "io_bazel_rules_docker",
    remote = "https://github.com/bazelbuild/rules_docker.git",
    tag = "v0.3.0",
)

load(
    "@io_bazel_rules_docker//container:container.bzl",
    "container_pull",
    container_repositories = "repositories",
)

container_repositories()

container_pull(
  name = "debian_hyperkube_base_amd64",
  registry = "gcr.io",
  repository = "google_containers/debian-hyperkube-base-amd64",
  # 'tag' is also supported, but digest is encouraged for reproducibility.
  digest = "sha256:fc1b461367730660ac5a40c1eb2d1b23221829acf8a892981c12361383b3742b",
  tag = "0.8",
)
