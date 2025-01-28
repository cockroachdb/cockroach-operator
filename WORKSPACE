workspace(name = "com_github_coachroachdb_cockroach_operator")

#####################################
# Bazel macros for downloading deps #
#####################################
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

http_archive(
    name = "com_google_protobuf",
    sha256 = "b07772d38ab07e55eca4d50f4b53da2d998bb221575c60a4f81100242d4b4889",
    strip_prefix = "protobuf-3.20.0",
    urls = ["https://github.com/protocolbuffers/protobuf/archive/v3.20.0.tar.gz"],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

http_archive(
    name = "com_github_bazelbuild_buildtools",
    sha256 = "a02ba93b96a8151b5d8d3466580f6c1f7e77212c4eb181cba53eb2cae7752a23",
    strip_prefix = "buildtools-3.5.0",
    urls = [
        "https://github.com/bazelbuild/buildtools/archive/3.5.0.tar.gz",
    ],
)

#################################
# External Go Rules and Gazelle #
#################################

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "8e968b5fcea1d2d64071872b12737bbb5514524ee5f0a4f54f5920266c261acb",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.28.0/rules_go-v0.28.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.28.0/rules_go-v0.28.0.zip",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "62ca106be173579c0a167deb23358fdfe71ffa1e4cfdddf5582af26520f1c66f",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.23.0/bazel-gazelle-v0.23.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.23.0/bazel-gazelle-v0.23.0.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

# we have to log go_dependencies before gazelle because of
# and old version of http2 in the k8s API
load("//hack/build:repos.bzl", "go_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

go_rules_dependencies()

go_register_toolchains(version = "1.17")

gazelle_dependencies()

# gazelle:repository_macro hack/build/repos.bzl%_go_dependencies
go_dependencies()

################################
# begin rules_oci dependencies #
################################
http_archive(
    name = "aspect_bazel_lib",
    sha256 = "d0529773764ac61184eb3ad3c687fb835df5bee01afedf07f0cf1a45515c96bc",
    strip_prefix = "bazel-lib-1.42.3",
    url = "https://storage.googleapis.com/public-bazel-artifacts/bazel/bazel-lib-v1.42.3.tar.gz",
)

http_archive(
    name = "rules_oci",
    sha256 = "21a7d14f6ddfcb8ca7c5fc9ffa667c937ce4622c7d2b3e17aea1ffbc90c96bed",
    strip_prefix = "rules_oci-1.4.0",
    url = "https://storage.googleapis.com/public-bazel-artifacts/bazel/rules_oci-v1.4.0.tar.gz",
)

load("@rules_oci//oci:dependencies.bzl", "rules_oci_dependencies")
load("@rules_oci//oci:pull.bzl", "oci_pull")

rules_oci_dependencies()

# TODO: This will pull from an upstream location: specifically it will download
# `crane` from https://github.com/google/go-containerregistry/... Before this is
# used in CI or anything production-ready, this should be mirrored. rules_oci
# doesn't support this mirroring yet so we'd have to submit a patch.
load("@rules_oci//oci:repositories.bzl", "LATEST_CRANE_VERSION", "oci_register_toolchains")

oci_register_toolchains(
    name = "oci",
    crane_version = LATEST_CRANE_VERSION,
)

load("@aspect_bazel_lib//lib:repositories.bzl", "aspect_bazel_lib_dependencies")

aspect_bazel_lib_dependencies()

##############################
# end rules_oci dependencies #
##############################

oci_pull(
    name = "redhat_ubi_minimal",
    platforms = [
        "linux/amd64",
        "linux/arm64",
    ],
    registry = "registry.access.redhat.com",
    repository = "ubi8/ubi-minimal",
    tag = "latest",
)

http_archive(
    name = "rules_pkg",
    sha256 = "8f9ee2dc10c1ae514ee599a8b42ed99fa262b757058f65ad3c384289ff70c4b8",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_pkg/releases/download/0.9.1/rules_pkg-0.9.1.tar.gz",
        "https://github.com/bazelbuild/rules_pkg/releases/download/0.9.1/rules_pkg-0.9.1.tar.gz",
    ],
)

load("@rules_pkg//:deps.bzl", "rules_pkg_dependencies")

rules_pkg_dependencies()

###################################################
# Load kubernetes repo-infra for tools like kazel #
###################################################
http_archive(
    name = "io_k8s_repo_infra",
    sha256 = "ae75a3a8de9698df30dd5a177c61f31ae9dd3a5da96ec951f0d6e60b2672d5fe",
    strip_prefix = "repo-infra-0.2.2",
    urls = [
        "https://github.com/kubernetes/repo-infra/archive/v0.2.2.tar.gz",
    ],
)

#################################################
# Load and define targets defined in //hack/bin #
#################################################
load("//hack/bin:deps.bzl", install_hack_bin = "install")

install_hack_bin()
