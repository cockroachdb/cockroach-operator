# Install Cockroach Operator on an Openshift cluster

1. Create a new image on a custom repo:

```
 make release/image
```

2. Create bundle:
```
make release/update-pkg-manifest
make release/opm-build-bundle
```

3. Release bundle
```
```

4. Install operator from Operator Hub
```
oc delete crd crdbclusters.crdb.cockroachlabs.com
kubectl delete pvc,pv  --all
oc delete -f test-registry.yaml
oc apply -f test-registry.yaml
```