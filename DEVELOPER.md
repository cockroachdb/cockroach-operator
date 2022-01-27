# Developer Notes

This project requires bazel. Version 4.0.0 has been tested.

See: https://docs.bazel.build/versions/master/install.html

Take a look at the [Makefile](https://github.com/cockroachdb/cockroach-operator/blob/master/Makefile) for various
targets you can run. The e2e testing requires a container engine like docker running as it starts a K8s cluster with
[k3d]. The Makefile is simply a wrapper for bazel commands and does not have any build functionality in it.

You will also need kubectl and a running Kubernetes cluster to run the k8s targets like `make k8s/apply`.

Bazel has been primarily tested on Linux but "should" run on macOS as well. Windows at this time has not been tested.

If you are new to Bazel take a look at the Workspace file, as it lays out which rule sets we are using, such as docker
rules and k8s rules. The Makefile notes below talk about the different target groups that run various bazel commands.

### Requirements

Just a quick recap of above. Please install the following.

1. Bazel
1. Container Engine
1. kubectl

## Makefile Notes

We have four different groups of commands withing the Makefile

1. Testing targets - targets that a run during CI and you can run locally
2. Dev targets - build commands
3. Kubernetes targets - we use `rules_k8s` that allow the deployment of the operatator to an existing registry and k8s
   cluster
4. Release targets - this is WIP, but allows for pushing the operator container to a registry

Please run the testing targets locally before you push a PR.

## Running Locally

We use [k3d] to run the operator in a local environment. This can provide a faster feedback cycle while developing
since there's no need to set up a remote GKE/OpenShift cluster.

[k3d]: https://k3d.io

**make dev/up**

This command will get everything set up for you to begin testing out the operator. Specifically it will:

* Start a k3d cluster named test (context=k3d-test) with a managed docker registry
* Install the CRDs
* Build a docker image and publish it to the k3d registry
* Deploy the operator and wait for it to be available
* Ensure your shell is configured to use the new cluster (kube context)

**make dev/down**

Tears down the k3d cluster.

## Testing CR Database

Notes on how to test an existing CR Database.

kubectl exec into a pod

```
cockroach sql
CREATE TABLE bank.accounts (id INT PRIMARY KEY, balance DECIMAL);
INSERT INTO bank.accounts VALUES (1, 1000.50);
SELECT * FROM bank.accounts;
```

## Developer Install Instructions

These instructions are for developers only. If you want to try the alpha please use the instructions in the next
section.

Install the operator

```console
$ git clone https://github.com/cockroachdb/cockroach-operator.git
$ export CLUSTER=test
# create a gke cluster
$ ./hack/create-gke-cluster.sh -c $CLUSTER

$ DEV_REGISTRY=us.gcr.io/$(gcloud config get-value project) \
K8S_CLUSTER=$(kubectl config view --minify -o=jsonpath='{.contexts[0].context.cluster}') \
bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
                 //manifests:install_operator.apply
```

There are various examples that can be installed. The files are located in the examples directory.

## Delete the cluster

When you have removed your example and the persitent volumes you can use the following command to delete your cluster.

```bash
# delete the cluster
# note this is async, and the script will complete without waiting the entire time
./hack/delete-gke-cluster.sh -c $CLUSTER
```

## Dev Release Install Instructions

1. If you do not have a GKE cluster we have a helper script `./hack/create-gke-cluster.sh -c test`.
1. Locate the latest released tag here  https://github.com/cockroachdb/cockroach-operator/tags
1. Clone the tag `git clone --depth 1 --branch <tag_name> https://github.com/cockroachdb/cockroach-operator`
1. Execute `kubectl apply -f install/crds.yaml`
1. Execute `kubectl apply -f install/operator.yaml`
1. Check that the operator has deployed properly

The examples directory contains various examples, for example you can run `kubectl apply -f examples/example.yaml`.

## Accessing Admin UI

First you need to create a new Kubernetes Custom Resource, in this example we will use the example/example.yaml file.

```bash
kubectl create -f example/example.yaml
```

When the database is up and running run the following command to get the first pod that is creasted.

```bash
kubectl get po -l app.kubernetes.io/instance=crdb-example --no-headers=true | head -n 1 | awk '{ print $1 }'
```

The command-line above uses "crdb-example" as it is the instance name in example.yaml file. If you name your database
differently then use the appropriate name. The command in the above example will output the name `crdb-example-0`, and
use that name in the next command.

```bash
kubectl port-forward crdb-example-0 8080
```

Then point your browser to http://localhost:8080 and you will load the admin UI.

NOTE: If you are installing the database in a namespace other than default, please add the `-n` flag appropriately.

## Install Cockroach Operator on an Openshift cluster

1. Create a new image on a custom repo:

```bash
 make release/image
```

2. Create bundle:

```bash
make release/update-pkg-bundle
```

3. Release bundle -- To Do

4. Cleanup existing cluster and register new version on OperatorHub

```bash
oc delete crd crdbclusters.crdb.cockroachlabs.com
kubectl delete pvc,pv  --all
oc delete -f test-registry.yaml
oc apply -f test-registry.yaml
```

5. Install operator from Operator Hub

## Generated configs

Some configs contain information about CockroachDB versions, for example docker image that contains CRDB version. In
order to keep those values up to date, we generate some configs using templates. The templates can be found in the
`config/templates` directory.

### How to add a new template

* Create a template file using [Go templates](https://pkg.go.dev/text/template). The variables for the templates are
  defined as `templateData` in
  `hack/crdbversions/main.go`.
* Add the new template to the `targets` variable in `hack/crdbversions/main.go`.
* Regenerate the outputs by running `make release/gen-templates`
