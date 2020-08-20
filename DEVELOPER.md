# Developer Notes

This project requires bazel. Version 3.4.1 has been tested.

See: https://docs.bazel.build/versions/master/install.html

Take a look at the [](Makefile) for various targets you can run. The e2e testing requires a container engine like docker running
as it starts a K8s cluster with Kind.  The Makefile is simply a wrapper for bazel commands and does not
have any build functionality in it.

You will also need kubectl and a running Kubernetes cluster to run the k8s targets like `make k8s/apply`. 

Bazel has been primarily tested on Linux but "should" run on macOS as well.  Windows at this time has not been tested.

TODO: notes on python configuration for the host.

If you are new to Bazel take a look at the Workspace file, as it lays out which rule sets we are using, such as docker rules and k8s rules.  The Makefile notes below talk about the different target groups that run various bazel commands.

### Requirements

Just a quick recap of above.  Please install the following.

1. Bazel
1. Python - TODO version 
1. Container Engine
1. kubectl

If you are not running the e2e tests or the Kubernetes targets you not need the last two.  But you probably have them anyways.


## Makefile Notes

We have four different groups of commands withing the Makefile

1. Testing targets - targets that a run during CI and you can run locally
2. Dev targets - build commands
3. Kubernetes targets - we use `rules_k8s` that allow the deployment of the operatator to an existing registry and k8s cluster
4. Release targets - this is WIP, but allows for pushing the operator container to a registry

Please run the testing targets locally before you push a PR.


## Testing CR Database

Notes on how to test an existing CR Database.

kubectl exec into a pod

```
cockroach sql
CREATE TABLE bank.accounts (id INT PRIMARY KEY, balance DECIMAL);
INSERT INTO bank.accounts VALUES (1, 1000.50);
SELECT * FROM bank.accounts;
```

## Running the operator in GCP

Install the operator

```
git clone https://github.com/cockroachdb/cockroach-operator.git
export CLUSTER=test
# create a gke cluster
./hack/create-gke-cluster.sh -c $CLUSTER

export IMAGE_REGISTRY=us.gcr.io/$(gcloud config get-value project)
export K8S_CLUSTER=$(kubectl config view --minify -o=jsonpath='{.contexts[0].context.cluster}')
bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
                 //manifests:install_operator.apply
```

There are various examples that can be installed.  The files are located in the examples directory. You can also use:

```
# install a basic example
./hack/apply-crdb-example.sh -c $CLUSTER
 ```

Delete the cluster

```
# delete the example
./hack/delete-crdb-example.sh -c $CLUSTER

# If you're still using the gke cluster, you can delete persistent volumes and persistent volume claims. It is not recommended to do this in production. Use --help for details.
kubectl delete pv,pvc --help

# delete the cluster
# note this is async, and the script will complete without waiting the entire time
./hack/delete-gke-cluster.sh -c $CLUSTER
```
