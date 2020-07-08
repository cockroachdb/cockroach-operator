# cockroach-operator 
Kubernetes Operator for CockroachDB

This project is not production ready and is in an alpha state.

# Development

### Requirements

- Docker

The rest of dependencies are packed into Docker images which are executed from the [](Makefile)

### Running the operator in GCP

```
git clone https://github.com/cockroachdb/cockroach-operator.git
export CLUSTER=test
# create a gke cluster
./hack/create-gke-cluster.sh -t $CLUSTER

# build the image locally and push it to your image repo
# record the output of the command, as it will tell you the image to use with kustomize
./hack/push-operator-gcr.sh

# apply the operator
# replace the -i argument with the image name
./hack/apply-operator.sh -i myimagename:2342 -c $CLUSTER

# validate the the operator is running
alias kubectl=k
k get po

# install a basic example
./hack/apply-crdb-example.sh -c $CLUSTER
```
You now should have a basic crdb running. You can kubectl exec into a pod and test the database.

Clean up the cluster

```
# delete the example
./hack/delete-crdb-example.sh -c $CLUSTER

# delete the operator
./hack/delete-operator.sh -c $CLUSTER

# delete the cluster
# note this is async, and the script will complete without waiting the entire time
./hack/delete-gke-cluster.sh -c $CLUSTER
```

### Running operator in local development environment

If you have kubectl configured to use a cluster, you can start the operator locally:

```
# Install CRD (you may need to add --validate=false for k8s version below 1.15
kubectl apply -f ./config/crd/bases/crdb.cockroachlabs.com_crdbclusters.yaml

# Start operator - run the operator in local development mode
WATCH_NAMESPACE=default
make run
```

### Recipes:

The main [](Makefile) provides a set of targets to automate development and deployment. Here are the most useful:

- `make run` runs the operator which attempts to connect to the current cluster from kubectl config
- `make test` runs tests for the project
- `make manifests` generates manifests (e.g. CRD, RBAC) from Go types
- `make generate` generates Go code from Go types
