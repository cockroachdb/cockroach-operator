# CockroachDB Kubernetes Operator

The CockroachDB Kubernetes Operator deploys CockroachDB on a Kubernetes cluster. You can use the Operator to manage the configuration of a running CockroachDB cluster, including:

- Authenticating certificates
- Configuring resource requests and limits
- Scaling the cluster
- Performing a rolling upgrade

## Limitations

- This project is in alpha and is not production-ready.
- The Operator currently runs on GKE.
- Currently the Kubernetes CA is required to deploy a secure cluster.
- Expansion of persistent volumes is not yet functional.

## Prerequisites

- Kubernetes 1.8 or higher
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- A GKE cluster

## Install the Operator

Locate the latest release tag [here](https://github.com/cockroachdb/cockroach-operator/tags).

Clone the release tag:

```
git clone --depth 1 --branch <tag_name> https://github.com/cockroachdb/cockroach-operator
```

Apply the CustomResourceDefinition (CRD) for the Operator:

```
kubectl apply -f config/crd/bases/crdb.cockroachlabs.com_crdbclusters.yaml
```

Apply the Operator manifest:

```
kubectl apply -f manifests/operator.yaml
```

Validate that the Operator is running:

```
kubectl get pods
```

```
NAME                                  READY   STATUS    RESTARTS   AGE
cockroach-operator-6f7b86ffc4-9ppkv   1/1     Running   0          54s
```

## Start CockroachDB

Download [`example.yaml`](https://github.com/cockroachdb/cockroach-operator/examples/example.yaml) and optionally modify the default configuration as outlined below.

```
apiVersion: crdb.cockroachlabs.com/v1alpha1
kind: CrdbCluster
metadata:
  name: cockroachdb
spec:
  dataStore:
    pvc:
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
          #   cpu: "4"
          #   memory: "16Gi"
            storage: 60Gi
          # limits:
          #   cpu: "4"
          #   memory: "16Gi"
        volumeMode: Filesystem
  tlsEnabled: true
  # nodeTLSSecret: cockroachdb.node
  # clientTLSSecret: cockroachdb.client.root
  image:
    name: cockroachdb/cockroach:v20.1.4
  nodes: 3
```

### Certificate signing

By default, the Operator uses the built-in Kubernetes CA to generate and approve 1 client and 1 node certificate for the cluster.

<!-- If you wish to use a non-Kubernetes CA, additionally enable the following lines in `example.yaml` and fill in the names of the corresponding node and client secrets:

```
nodeTLSSecret: <node-secret>
clientTLSSecret: <client-secret>
```

For example, if you are using [`cockroach cert`](https://www.cockroachlabs.com/docs/stable/cockroach-cert.html) to generate certificates, use these secret names:

```
nodeTLSSecret: cockroachdb.node
clientTLSSecret: cockroachdb.client.root
``` -->

### Set resource requests and limits

Before deploying, we recommend explicitly allocating CPU and memory to CockroachDB in the Kubernetes pods. 

Enable the commented-out lines in the `resources.requests` object and substitute values appropriate for your workload. Note that `requests` and `limits` should have identical values.

```
resources:
  requests:
    cpu: "4"
    memory: "16Gi"
    storage: 60Gi
  limits:
    cpu: "4"
    memory: "16Gi"
```

### Apply the StatefulSet configuration

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

Get a shell into one of the pods and start the CockroachDB built-in SQL client:

```
kubectl exec -it cockroachdb-2 -- ./cockroach sql --certs-dir=/cockroach-certs
```

If you want to [access the Admin UI](#access-the-admin-ui), create a SQL user with a password while you're here:

```
CREATE USER roach WITH PASSWORD 'Q7gc8rEdS';
```

```
\q
```

## Access the Admin UI

To access the cluster's [Admin UI](https://www.cockroachlabs.com/docs/v20.2/admin-ui-overview), port-forward from your local machine to the `cockroachdb-public` service:

```
kubectl port-forward service/cockroachdb-public 8080
```

Access the Admin UI at `https://localhost:8080`. Note that [certain Admin UI pages](https://www.cockroachlabs.com/docs/v20.2/admin-ui-overview#admin-ui-access) require `admin` access, which you can grant to the SQL user you created earlier:

```
GRANT admin TO roach;
```

### Scale the cluster

To scale the cluster up and down, edit `nodes` in `example.yaml`:

```
nodes: 4
```

Do **not** scale down to fewer than 3 nodes. This is considered an anti-pattern on CockroachDB and will cause errors.

> Note that you must scale by updating the `nodes` value in the Operator configuration. Using `kubectl scale statefulset <cluster-name> --replicas=4` will result in new pods immediately being terminated.

Apply the new configuration:

```
kubectl apply -f example.yaml
```

### Upgrade the cluster

Perform a rolling upgrade by changing the image name in `example.yaml`:

```
image:
  name: cockroachdb/cockroach:v20.2.0
```

Apply the new configuration:

```
kubectl apply -f example.yaml
```

## Stop the cluster

Delete the StatefulSet:

```
kubectl delete -f example.yaml
```

Remove the Operator:

```
kubectl delete -f cockroach-operator/manifests/operator.yaml
```