# Copyright 2019 The Jetstack cert-manager contributors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file", "http_archive")
load("@io_bazel_rules_docker//container:container.bzl", "container_pull")
load("@bazel_gazelle//:deps.bzl", "go_repository")

def install():
    install_misc()
    install_integration_test_dependencies()
    install_bazel_tools()
    install_staticcheck()
    #install_helm()
    install_kubectl()
    install_oc()
    install_kind()
    install_kubetest2()
    install_kubetest2_kind()
    install_kubetest2_gke()
    install_kubetest2_exe()
    install_operator_sdk()
    install_kustomize()
    install_opm()
    install_crdb()
    install_openshift()
    install_aws_kubetest2()

    # Install golang.org/x/build as kubernetes/repo-infra requires it for the
    # build-tar bazel target.
    go_repository(
        name = "org_golang_x_build",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/build",
        sum = "h1:hXVePvSFG7tPGX4Pwk1d10ePFfoTCc0QmISfpKOHsS8=",
        version = "v0.0.0-20190927031335-2835ba2e683f",
    )

def install_staticcheck():
    http_archive(
        name = "co_honnef_go_tools_staticcheck_linux",
        sha256 = "09d2c2002236296de2c757df111fe3ae858b89f9e183f645ad01f8135c83c519",
        urls = ["https://github.com/dominikh/go-tools/releases/download/2020.1.4/staticcheck_linux_amd64.tar.gz"],
        build_file_content = """
filegroup(
    name = "file",
    srcs = [
        "staticcheck/staticcheck",
    ],
    visibility = ["//visibility:public"],
)
""",
    )

    http_archive(
        name = "co_honnef_go_tools_staticcheck_osx",
        sha256 = "5706d101426c025e8f165309e0cb2932e54809eb035ff23ebe19df0f810699d8",
        urls = ["https://github.com/dominikh/go-tools/releases/download/2020.1.4/staticcheck_darwin_amd64.tar.gz"],
        build_file_content = """
filegroup(
    name = "file",
    srcs = [
        "staticcheck/staticcheck",
    ],
    visibility = ["//visibility:public"],
)
""",
    )

def install_misc():
    http_file(
        name = "jq_linux",
        executable = 1,
        sha256 = "c6b3a7d7d3e7b70c6f51b706a3b90bd01833846c54d32ca32f0027f00226ff6d",
        urls = ["https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64"],
    )

    http_file(
        name = "jq_osx",
        executable = 1,
        sha256 = "386e92c982a56fe4851468d7a931dfca29560cee306a0e66c6a1bd4065d3dac5",
        urls = ["https://github.com/stedolan/jq/releases/download/jq-1.5/jq-osx-amd64"],
    )
    http_file(
        #this tool is used on open shift generation to generate the csv
        name = "faq_linux",
        executable = 1,
        sha256 = "53360a0d22b0608d5e29f8e84450f2fdc94573246fb552896afedbf8f1687981",
        urls = ["https://github.com/jzelinskie/faq/releases/download/0.0.6/faq-linux-amd64"],
    )

    http_file(
        name = "faq_osx",
        executable = 1,
        sha256 = "bfcd6f527d1ba74db6bdd6bfb551a4db9c2c72f01baebf8069e9849b93dceef9",
        urls = ["https://github.com/jzelinskie/faq/releases/download/0.0.6/faq-darwin-amd64"],
    )

# Install dependencies used by the controller-runtime integration test framework
def install_integration_test_dependencies():
    http_file(
        name = "kube-apiserver_darwin_amd64",
        executable = 1,
        sha256 = "a874d479f183f9e4c19a5c69b44955fabd2e250b467d2d9f0641ae91a82ddbea",
        urls = ["https://storage.googleapis.com/cert-manager-testing-assets/kube-apiserver-1.17.3_darwin_amd64"],
    )

    http_file(
        name = "kube-apiserver_linux_amd64",
        executable = 1,
        sha256 = "b4505b838b27b170531afbdef5e7bfaacf83da665f21b0e3269d1775b0defb7a",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.17.3/bin/linux/amd64/kube-apiserver"],
    )

    http_archive(
        name = "com_coreos_etcd_darwin_amd64",
        sha256 = "c8f36adf4f8fb7e974f9bafe6e390a03bc33e6e465719db71d7ed3c6447ce85a",
        urls = ["https://github.com/etcd-io/etcd/releases/download/v3.3.12/etcd-v3.3.12-darwin-amd64.zip"],
        build_file_content = """
filegroup(
    name = "file",
    srcs = [
        "etcd-v3.3.12-darwin-amd64/etcd",
    ],
    visibility = ["//visibility:public"],
)
""",
    )

    http_archive(
        name = "com_coreos_etcd_linux_amd64",
        sha256 = "dc5d82df095dae0a2970e4d870b6929590689dd707ae3d33e7b86da0f7f211b6",
        urls = ["https://github.com/etcd-io/etcd/releases/download/v3.3.12/etcd-v3.3.12-linux-amd64.tar.gz"],
        build_file_content = """
filegroup(
    name = "file",
    srcs = [
        "etcd-v3.3.12-linux-amd64/etcd",
    ],
    visibility = ["//visibility:public"],
)
""",
    )

# Install additional tools for Bazel management
def install_bazel_tools():
    ## Install buildozer, for mass-editing BUILD files
    http_file(
        name = "buildozer_darwin",
        executable = 1,
        sha256 = "f2bcb59b96b1899bc27d5791f17a218f9ce76261f5dcdfdbd7ad678cf545803f",
        urls = ["https://github.com/bazelbuild/buildtools/releases/download/0.22.0/buildozer.osx"],
    )

    http_file(
        name = "buildozer_linux",
        executable = 1,
        sha256 = "7750fe5bfb1247e8a858f3c87f63a5fb554ee43cb10efc1ce46c2387f1720064",
        urls = ["https://github.com/bazelbuild/buildtools/releases/download/0.22.0/buildozer"],
    )

# Install Helm targets
def install_helm():
    ## Fetch helm & tiller for use in template generation and testing
    ## You can bump the version of Helm & Tiller used during e2e tests by tweaking
    ## the version numbers in these rules.
    http_archive(
        name = "helm_darwin",
        sha256 = "92b10652b05a150e76995e08910a662c200a8179cfdb16bd51766d0d5ecc981a",
        urls = ["https://get.helm.sh/helm-v3.1.2-darwin-amd64.tar.gz"],
        build_file_content =
            """
filegroup(
    name = "file",
    srcs = [
        "darwin-amd64/helm",
    ],
    visibility = ["//visibility:public"],
)
""",
    )

    http_archive(
        name = "helm_linux",
        sha256 = "e6be589df85076108c33e12e60cfb85dcd82c5d756a6f6ebc8de0ee505c9fd4c",
        urls = ["https://get.helm.sh/helm-v3.1.2-linux-amd64.tar.gz"],
        build_file_content =
            """
filegroup(
    name = "file",
    srcs = [
        "linux-amd64/helm",
    ],
    visibility = ["//visibility:public"],
)
""",
    )

# Define rules for different kubectl versions
def install_kubectl():
    http_file(
        name = "kubectl_1_18_darwin",
        executable = 1,
        sha256 = "5eda86058a3db112821761b32afce3fdd2f6963ab580b1780a638ac323864eba",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.18.0/bin/darwin/amd64/kubectl"],
    )

    http_file(
        name = "kubectl_1_18_linux",
        executable = 1,
        sha256 = "bb16739fcad964c197752200ff89d89aad7b118cb1de5725dc53fe924c40e3f7",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.18.0/bin/linux/amd64/kubectl"],
    )


# Define rules for different oc versions
def install_oc():
    http_archive(
        name = "oc_linux",
        sha256 = "d3b62b7de6df34b75336fc4720e334b8073cc93dba09c07f2e6a8e7147c0c99a",
        urls = ["https://mirror.openshift.com/pub/openshift-v4/clients/ocp/4.7.15/openshift-client-linux-4.7.15.tar.gz"],
        build_file_content =
         """
filegroup(
     name = "file",
     srcs = [
        "oc",
     ],
     visibility = ["//visibility:public"],
)
    """,
    )

    http_archive(
        name = "oc_darwin",
        sha256 = "835218ded58cdda11f83c7e777a79a36c97c30c00c72f28a5ee29d8fb7e8830e",
        urls = ["https://mirror.openshift.com/pub/openshift-v4/clients/ocp/4.7.15/openshift-client-mac-4.7.15.tar.gz"],
        build_file_content =
         """
filegroup(
     name = "file",
     srcs = [
        "oc",
     ],
     visibility = ["//visibility:public"],
)
    """,
    )

## Fetch kind images used during e2e tests
def install_kind():
    # install kind binary
    http_file(
        name = "kind_darwin",
        executable = 1,
        sha256 = "432bef555a70e9360b44661c759658265b9eaaf7f75f1beec4c4d1e6bbf97ce3",
        urls = ["https://github.com/kubernetes-sigs/kind/releases/download/v0.11.1/kind-darwin-amd64"],
    )

    http_file(
        name = "kind_linux",
        executable = 1,
        sha256 = "949f81b3c30ca03a3d4effdecda04f100fa3edc07a28b19400f72ede7c5f0491",
        urls = ["https://github.com/kubernetes-sigs/kind/releases/download/v0.11.1/kind-linux-amd64"],
    )

## Fetch kubetest2 binary used during e2e tests
def install_kubetest2():
    # install kubetest2 binary
    # TODO osx support
    http_file(
       name = "kubetest2_darwin",
       executable = 1,
       sha256 = "5b20aadd05eca47dead180a7c8296d75e81c184aabf182d4a41ef96597db543d",
       urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/osx/kubetest2"],
    )

    http_file(
        name = "kubetest2_linux",
        executable = 1,
        sha256 = "7f0b05654fa43ca1c607db297b5f3a775f65eea90355bb6b10137a7fffff5e1a",
        urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/linux/kubetest2"],
    )

## Fetch kubetest2-kind binary used during e2e tests
def install_kubetest2_kind():
    # install kubetest2-kind binary
    # TODO osx support
    http_file(
       name = "kubetest2_kind_darwin",
       executable = 1,
       sha256 = "a68bad1b94fd5e432f0555d699d0ce0470d0bf16f1b087e857d55f16f5373385",
       urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/osx/kubetest2-kind"],
    )

    http_file(
        name = "kubetest2_kind_linux",
        executable = 1,
        sha256 = "b13014d3e1464ce58e2bbbec94bead267936155d537f3232ec0a24727263a2a1",
        urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/linux/kubetest2-kind"],
    )

## Fetch kubetest2-gke binary used during e2e tests
def install_kubetest2_gke():
    # install kubetest2-gke binary
    # TODO osx support
    http_file(
       name = "kubetest2_gke_darwin",
       executable = 1,
       sha256 = "a1cbe02f61931dbe6c8d1662442f42cb538c81e4ec8cdd40f548f0e05cbd55a7",
       urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/osx/kubetest2-gke"],
    )

    http_file(
        name = "kubetest2_gke_linux",
        executable = 1,
        sha256 = "9ac658234efc7f59968888662dd2d21908587789f6b812392ac5b6766b17c0b4",
        urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/linux/kubetest2-gke"],
    )
## Fetch kubetest2-tester-exe binary used during e2e tests
def install_kubetest2_exe():
    # install kubetest2-exe binary
    # TODO osx support
    http_file(
       name = "kubetest2_exe_darwin",
       executable = 1,
       sha256 = "818690cb55590440e163b18dd139c8a8714df9480f869bafe19eb344047cf37c",
       urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/osx/kubetest2-tester-exec"],
    )

    http_file(
        name = "kubetest2_exe_linux",
        executable = 1,
        sha256 = "4483f40f48b98e8a6aa41f58bfdf1f2787066a4e1ad1343e4281892aa1326736",
        urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/linux/kubetest2-tester-exec"],
    )
## Fetch operator-sdk used on generating csv
def install_operator_sdk():
    # install kubetest2-kind binary
    # TODO osx support
    http_file(
       name = "operator_sdk_darwin",
       executable = 1,
       sha256 = "7e293cd35b99c1949cbb116275bc50d7f3aa0b520fe7e53e57b09e8096e63d4e",
       urls = ["https://github.com/operator-framework/operator-sdk/releases/download/v1.1.0/operator-sdk-v1.1.0-x86_64-apple-darwin"],
    )

    http_file(
        name = "operator_sdk_linux",
        executable = 1,
        sha256 = "e0cfd1408ea8849fb32345d7f9954a2751fef7fcf4505f93db8f675d12f137ad",
        urls = ["https://github.com/operator-framework/operator-sdk/releases/download/v1.1.0/operator-sdk-v1.1.0-x86_64-linux-gnu"],
    )

     ## Fetch opm used on generating csv
def install_kustomize():
    http_archive(
       name = "kustomize_darwin",
       sha256 = "77898f8b7c37e3ba0c555b4b7c6d0e3301127fa0de7ade6a36ed767ca1715643",
       urls = ["https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv4.3.0/kustomize_v4.3.0_darwin_amd64.tar.gz"],
       build_file_content = """
filegroup(
    name = "file",
    srcs = [
	"kustomize",
    ],
    visibility = ["//visibility:public"],
)
""",
    )

    http_archive(
        name = "kustomize_linux",
        sha256 = "d34818d2b5d52c2688bce0e10f7965aea1a362611c4f1ddafd95c4d90cb63319",
        urls = ["https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv4.3.0/kustomize_v4.3.0_linux_amd64.tar.gz"],
        build_file_content = """
filegroup(
    name = "file",
    srcs = [
	"kustomize",
    ],
    visibility = ["//visibility:public"],
)
""",
    )

## Fetch opm used on generating csv
def install_opm():
    http_file(
       name = "opm_darwin",
       executable = 1,
       sha256 = "adbfcdeef14c9a5a8ebfc1e94d57f3c3892c85477873187bfb65ef226d757a9a",
       urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/macos/opm"],
    )

    http_file(
        name = "opm_linux",
        executable = 1,
        sha256 = "e3f15fbad17c903c7d69579e934153cb74fbd48ba84e4911d2ecf4e63b9a903d",
        urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/linux/opm"],
    )

## Fetch openshift-installer
def install_openshift():
    http_archive(
       name = "openshift_darwin",
       sha256 = "ed2144b31aa15bd27cbea11c3e94f7329531775b13b0b1dabb1c0ffa82ca4dc6",
       urls = ["https://mirror.openshift.com/pub/openshift-v4/clients/ocp/4.7.9/openshift-install-mac-4.7.9.tar.gz"],
       build_file_content = """
filegroup(
    name = "file",
    srcs = [
	"openshift-install",
    ],
    visibility = ["//visibility:public"],
)
""",
    )

    http_archive(
        name = "openshift_linux",
        sha256 = "a41682490aabdb6e0d0ca608f76c91b9db5c4b295febc9d6b9182932d5e34760",
        urls = ["https://mirror.openshift.com/pub/openshift-v4/clients/ocp/4.7.9/openshift-install-linux-4.7.9.tar.gz"],
        build_file_content = """
filegroup(
    name = "file",
    srcs = [
	"openshift-install",
    ],
    visibility = ["//visibility:public"],
)
""",
    )

## Fetch crdb used in our container
def install_crdb():
    http_archive(
       name = "crdb_darwin", # todo fix or remove
       sha256 = "bbe3a03c661555e8b083856c56c8a3b459f83064d1d552ed3467cbfb66e76db7",
       urls = ["https://binaries.cockroachdb.com/cockroach-v20.2.5.darwin-10.9-amd64.tgz"],
       build_file_content = """
filegroup(
    name = "file",
    srcs = [
        "cockroach-v20.2.5.darwin-10.9-amd64/cockroach",
    ],
    visibility = ["//visibility:public"],
)
""",
   )

    http_archive(
        name = "crdb_linux",
        sha256 = "57f4b00c736d8511328d6f33997a3a66cb4ec7142cb126d872dade399a0922e6",
        urls = ["https://binaries.cockroachdb.com/cockroach-v20.2.5.linux-amd64.tgz"],
        build_file_content = """
filegroup(
    name = "file",
    srcs = [
        "cockroach-v20.2.5.linux-amd64/cockroach",
    ],
    visibility = ["//visibility:public"],
)
""",
   )

# fetch binaries for aws kubetest2
def install_aws_kubetest2():
    http_file(
        name = "aws-k8s-tester-linux",
        executable = 1,
        sha256 = "f6a95feef94ab9a96145fff812eeebae14f974edfead98046351f3098808df54",
        urls = ["https://github.com/aws/aws-k8s-tester/releases/download/v1.6.1/aws-k8s-tester-v1.6.1-linux-amd64"],
    )
    http_file(
        name = "aws-k8s-tester-darwin",
        executable = 1,
        sha256 = "a724288e8f2b87df89c711ed59fa2a09db5ad2b50a35cb3039fa610408b99b32",
        urls = ["https://github.com/aws/aws-k8s-tester/releases/download/v1.6.1/aws-k8s-tester-v1.6.1-darwin-amd64"],
    )
