
## How to manually install on Quay a new version of the bundle for DEV purpose

1. Make sure the env var from the Makefile which start with RH are correctly set (operator version, bundle versions etc)

2. Run the following commands to create the bundle image:

```bash
make gen-csv

docker build -t quay.io/alinalion/cockroachdb-operator-bundle:1.2.25  -f bundle.Dockerfile .
docker push quay.io/alinalion/cockroachdb-operator-bundle:1.2.25

opm index add --bundles  quay.io/alinalion/cockroachdb-operator-bundle:1.2.25 --tag  quay.io/alinalion/cockroachdb-operator-index:1.2.25 -c docker

docker build -t  quay.io/alinalion/cockroachdb-operator-bundle:1.2.25 -f bundle.Dockerfile .
docker push quay.io/alinalion/cockroachdb-operator-index:1.2.25
```

3. Replace in text-registry.yaml the PLACEHOLDER with: quay.io/alinalion/cockroachdb-operator-index:1.2.25

4. Run command for cleanup and install in operator hub:

```bash
oc delete crd crdbclusters.crdb.cockroachlabs.com
kubectl delete pvc,pv  --all

oc delete -f test-registry.yaml
oc apply -f test-registry.yaml
```

5. Now you can manually install the version from OperatorHub


### How to install a version on Quay that was generated for Certification pipeline(release):


1. Run command:

```bash
make release/update-pkg-manifest
make release/opm-build-bundle

cd deploy/certified-metadata-bundle/cockroach-operator 
docker build -t quay.io/alinalion/cockroachdb-operator-bundle:1.2.26  -f bundle.Dockerfile .
docker push quay.io/alinalion/cockroachdb-operator-bundle:1.2.26

cd ../../..
./opm index add --bundles  quay.io/alinalion/cockroachdb-operator-bundle:1.2.26 --tag  quay.io/alinalion/cockroachdb-operator-index:1.2.26 -c docker

cd deploy/certified-metadata-bundle/cockroach-operator 
docker build -t  quay.io/alinalion/cockroachdb-operator-bundle:1.2.26 -f bundle.Dockerfile .
docker push quay.io/alinalion/cockroachdb-operator-index:1.2.26

cd ../../..
```

2. Replace in text-registry.yaml the PLACEHOLDER with: quay.io/alinalion/cockroachdb-operator-index:1.2.26

```bash
oc delete crd crdbclusters.crdb.cockroachlabs.com
kubectl delete pvc,pv  --all
oc delete -f test-registry.yaml
oc apply -f test-registry.yaml

```