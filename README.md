# CockroachDB Kubernetes Operator

The CockroachDB Kubernetes Operator deploys CockroachDB on a Kubernetes cluster. You can use the Operator to manage the configuration of a running CockroachDB cluster, including:

- Authenticating certificates
- Configuring resource requests and limits
- Scaling the cluster
- Performing a rolling upgrade

## Build Status

GKE Nightly: [![GKE Nightly](https://teamcity.cockroachdb.com/guestAuth/app/rest/builds/buildType:Cockroach_CockroachOperator_Nightlies_GkeNightly/statusIcon)](https://teamcity.cockroachdb.com/viewType.html?buildTypeId=Cockroach_CockroachOperator_Nightlies_GkeNightly)

## Limitations

- This project is in beta and is not production-ready.
- The Operator currently runs on GKE. VMware Tanzu, EKS, and AKS have not been tested.
- The Operator does not yet support [multiple Kubernetes clusters for multi-region deployments](https://www.cockroachlabs.com/docs/stable/orchestrate-cockroachdb-with-kubernetes-multi-cluster.html#eks).
- [Migrating from a deployment using the Helm Chart to the Operator](https://github.com/cockroachdb/cockroach-operator/issues/140) has not been defined or tested.
- The Operator does not yet [automatically create an ingress object](https://github.com/cockroachdb/cockroach-operator/issues/76).
- The Operator has not been tested with [Istio](https://istio.io/).  

## Prerequisites

- Kubernetes 1.15 or higher (1.18 is recommended)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- A GKE cluster (`n2-standard-4` is the minimum requirement for testing)

## Install the Operator

Apply the custom resource definition (CRD) for the Operator:

```
kubectl apply -f https://raw.githubusercontent.com/cockroachdb/cockroach-operator/master/config/crd/bases/crdb.cockroachlabs.com_crdbclusters.yaml
```

Apply the Operator manifest. By default, the Operator is configured to install in the `default` namespace. To use the Operator in a custom namespace, download the Operator manifest and edit all instances of `namespace: default` to specify your custom namespace. Then apply this version of the manifest to the cluster with `kubectl apply -f {local-file-path}` instead of using the command below.

```
kubectl apply -f https://raw.githubusercontent.com/cockroachdb/cockroach-operator/master/manifests/operator.yaml
```

> **Note:** The Operator can only install CockroachDB into its own namespace. 

Validate that the Operator is running:

```
kubectl get pods
```

```
NAME                                  READY   STATUS    RESTARTS   AGE
cockroach-operator-6f7b86ffc4-9ppkv   1/1     Running   0          54s
```

## Start CockroachDB

Download the [`example.yaml`](https://github.com/cockroachdb/cockroach-operator/blob/master/examples/example.yaml) custom resource.

> **Note:** The latest stable CockroachDB release is specified by default in `image.name`.

### Resource requests and limits

By default, the Operator allocates 2 CPUs and 8Gi memory to CockroachDB in the Kubernetes pods. These resources are appropriate for `n2-standard-4` (GCP) and `m5.xlarge` (AWS) machines.

On a production deployment, you should modify the `resources.requests` object in the custom resource with values appropriate for your workload. For details, see the [CockroachDB documentation](https://www.cockroachlabs.com/docs/stable/operate-cockroachdb-kubernetes.html#allocate-resources).

### Certificate signing

The Operator generates and approves 1 root and 1 node certificate for the cluster.

### Apply the custom resource

Apply `example.yaml`:

```
kubectl create -f example.yaml
```

Check that the pods were created:

```
kubectl get pods
```

```
NAME                                  READY   STATUS    RESTARTS   AGE
cockroach-operator-6f7b86ffc4-9t9zb   1/1     Running   0          3m22s
cockroachdb-0                         1/1     Running   0          2m31s
cockroachdb-1                         1/1     Running   0          102s
cockroachdb-2                         1/1     Running   0          46s
```

Each pod should have `READY` status soon after being created.

## Access the SQL shell

To use the CockroachDB SQL client, first launch a secure pod running the `cockroach` binary.

```
kubectl create -f https://raw.githubusercontent.com/cockroachdb/cockroach-operator/master/examples/client-secure-operator.yaml
```

Get a shell into the client pod:

```
kubectl exec -it cockroachdb-client-secure -- ./cockroach sql --certs-dir=/cockroach/cockroach-certs --host=cockroachdb-public
```

If you want to [access the DB Console](#access-the-db-console), create a SQL user with a password while you're here:

```
CREATE USER roach WITH PASSWORD 'Q7gc8rEdS';
```

Then assign `roach` to the `admin` role to enable access to [secure DB Console pages](https://www.cockroachlabs.com/docs/stable/ui-overview.html#db-console-security):

```
GRANT admin TO roach;
```

```
\q
```

## Access the DB Console

To access the cluster's [DB Console](https://www.cockroachlabs.com/docs/stable/ui-overview.html), port-forward from your local machine to the `cockroachdb-public` service:

```
kubectl port-forward service/cockroachdb-public 8080
```

Access the DB Console at `https://localhost:8080`.

### Scale the CockroachDB cluster

To scale the cluster up and down, modify `nodes` in the custom resource. For details, see the [CockroachDB documentation](https://www.cockroachlabs.com/docs/stable/operate-cockroachdb-kubernetes#scale-the-cluster).

Do **not** scale down to fewer than 3 nodes. This is considered an anti-pattern on CockroachDB and will cause errors.

> **Note:** You must scale by updating the `nodes` value in the Operator configuration. Using `kubectl scale statefulset <cluster-name> --replicas=4` will result in new pods immediately being terminated.

### Upgrade the CockroachDB cluster

Perform a rolling upgrade by changing `image.name` in the custom resource. For details, see the [CockroachDB documentation](https://www.cockroachlabs.com/docs/stable/operate-cockroachdb-kubernetes#upgrade-the-cluster).

## Stop the CockroachDB cluster

Delete the custom resource:

```
kubectl delete -f example.yaml
```

Remove the Operator:

```
kubectl delete -f https://raw.githubusercontent.com/cockroachdb/cockroach-operator/master/manifests/operator.yaml
```

> **Note:** If you want to delete the persistent volumes and free up the storage used by CockroachDB, be sure you have a backup copy of your data. Data **cannot** be recovered once the persistent volumes are deleted. For more information, see the [Kubernetes documentation](https://kubernetes.io/docs/tasks/run-application/delete-stateful-set/#persistent-volumes). 

# Releases

The pre-release procedure requires you to adjust the version in `version.txt`
and run a command to generate the OpenShift metadata. See details below.

The version from the `version.txt` file is used for the following artifacts:

- Operator docker image published to Docker hub (using the "v" prefix, e.g. `v1.0.0`)
- Operator docker image published to RedHat Connect (using the "v" prefix, e.g. `v1.0.0`)
- OpenShift Metadata bundle (no "v" prefix, e.g. `1.0.0`), committed in the source directory
- OpenShift Metadata bundle docker image published to RedHat Connect (using the "v" prefix, e.g. `v1.0.0`)

## Bump the version in `version.txt`

- Create a PR branch in your local checkout, e.g. `git checkout -b release-1.0.0 origin/master`.
- Adjust the version in the `version.txt` file or set it from command line. Use
the semantic version schema (do not use the "v" prefix).
- Edit `Makefile` and adjust the `release/versionbump` target.
  - Set the `CHANNEL` to `beta` or `stable`.
  - Set the `DEFAULT_CHANNEL` to `0` or `1`.
- Run `make release/versionbump`.
- Push the local branch and request a review.
- After the PR is merged, tag the corresponding commit, e.g. `git tag v1.0.0 1234567890abcdef`.

## Run Release Automation
Release automation is run in TeamCity. This section will be updated after the
corresponding changes are merged.
