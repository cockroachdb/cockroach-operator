# OpenShift Kubetest2

This binary is a deployer that utilizes kubetest2 framework.
Kubetest2 is the framework for launching and running end-to-end tests on Kubernetes.
See https://github.com/kubernetes-sigs/kubetest2

This deployer will create a OpenShift cluster using openshift-install and then can
run tests using the new OpenShift cluster.  The cluster then can be torn down.
Only OpenShift on GCP is supported at this time.



## Deployer Details

`kubetest2` has various base flags, and for our purposes we only care about `--up`, `--down` and `--test`.
You can run `kubetest2 --help` for the full list of the flags. 

The OpenShift deployer uses various flags as well. The kubetest2 framework allows the addition of
more flags depending on which deployer or tester you are using.

To view these flags run `kubetest2 openshift --help`.

The flags that we are most concered about are:

```console
   --base-domain string               flag to set the TLD domain that is hosted in the GCP Project
   --cluster-name string              flag to set the name of the cluster
   --gcp-project-id string            flag for setting GCP project id, this needs to be the full id and not the common name.
   --gcp-region string                flag for setting GCP Region, for instance us-east1
   --pull-secret-file string          flag for the location of the pull secrete file downloaded from the RedHat site
   --sa-token string                  flag to set the path to the service account token
   --do-not-enable-services bool      flag to set to NOT enable services on the project
   --do-not-delete-sa bool            flag to set to NOT remove the service account
```

The `--base-domain` flag is set to the TLD domain that you want the `openshift-install` to use and must be hosted
in the same GCP project as defined via the `--gcp-project-id` flag.
A special note.  This deployer will create a service account that is used for running `openshift-install`. It does
use `gcloud` to create that IAM account, so your account must have the correct permissions.
The deployer will remove the IAM account as well.

Set the do-not flags if you are using an project that is already configured and has an existing service account. You will
need to set the sa-token flag to the json serice acount file.

## Setup and Building

### Build the deployer

First build this binary and download the other required binaries using bazel.

```console
bazel build //hack/bin/... //e2e/kubetest2-openshift/...
```

Bazel will download the kubetest2 binaries as well as the `openshift-install` binary.
These binaries are then placed in the `bazel-bin/hack/bin` and `bazel-bin/e2e/kubetest2-openshift/kubetest2-openshift_`
directories. Note, the gcloud binary is required and is not downloaded via bazel since it requires python.

### Setup your PATH

Export your path so that `kubetest2` can locate the binaries. You must do this command from the root
directory of this project in order for `pwd` to work.

```console
export PATH=${PATH}:$(pwd)/bazel-bin/hack/bin:$(pwd)/bazel-bin/e2e/kubetest2-openshift/kubetest2-openshift_/
```

## Examples

To create, runs tests, and destroy a cluster. Run this command:

```bash
kubetest2 openshift --cluster-name=${CLUSTER_NAME} \
  --gcp-project-id=${GCP_PROJECT} \
  --gcp-region=${GCP_REGION} \
  --base-domain=${BASE_DOMAIN} \
  --pull-secret-file=${PULL_SECRET} \
  --up --down \
  --test=exec -- <some command>
```

Note make sure you set the variables correctly and your $PATH correctly.
The above command will do the following actions.

1. `--up` will create the cluster
1. `--test=exec` will use the exec tester to execute whatever command your provide
1. `--down` will destroy the cluster

If you just want to create a cluster run:

```bash
kubetest2 openshift --cluster-name=${CLUSTER_NAME} \
  --gcp-project-id=${GCP_PROJECT} \
  --gcp-region=${GCP_REGION} \
  --base-domain=${BASE_DOMAIN} \
  --pull-secret-file=${PULL_SECRET} \
  --up
```

If you just want to delete the same cluster run:

```bash
kubetest2 openshift --cluster-name=${CLUSTER_NAME} \
  --gcp-project-id=${GCP_PROJECT_ID} --down
```

## Files Created

The OpenShift deployer creates and uses various files that are stored in your `$HOME` directory.

1. `$HOME/openhift-${CLUSTER_NAME}` is used to store various `openshift-install` files
1. `$HOME/.gcp-${CLUSTER_NAME}` is used to store the SA account token

When you do a `--down` option these directories and there contents are deleted.

## Troubleshooting

- When you are doing `--up` the above directories cannot exist.
- You need a TLD hosted via cloud dns in the GCP project that houses the cluster.
- If the deployer fails you may need to cleanup directories and possibly GCP resources such as the service account
- You need gcloud installed
- Your base account needs to have permission to create a SA with the correct roles or you need to provide a service account token location.

