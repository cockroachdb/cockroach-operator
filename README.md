# crdb-operator
k8s operator for CRDB

## Development

### Requirements

- GNU Make
- Docker
- go1.13

The rest of dependencies are packed into Docker images which are executed from the [](Makefile)

### Running operator in local development environment

If you have kubectl configured to use a cluster, you can start the operator locally:

```
# Install CRD (you may need to add --validate=false for k8s version below 1.15
kubectl apply -f ./config/crd/bases/crdb.cockroachlabs.com_crdbclusters.yaml

# Start operator
WATCH_NAMESPACE=default
make run
```

### Recipes:

The main [](Makefile) provides a set of targets to automate development and deployment. Here are the most useful:

- `make run` runs the operator which attempts to connect to the current cluster from kubectl config
- `make test` runs tests for the project
- `make manifests` generates manifests (e.g. CRD, RBAC) from Go types
- `make generate` generates Go code from Go types
