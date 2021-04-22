# Getting started with CockroachDB on Red Hat OpenShift

---

Today, we're going to cover the process of installing CockroachDB on the OpenShift platform. CockroachDB Kubernetes Operator is currently in development and this brief post will cover some of the available functionality to date.

## Prerequisites

1. CockroachDB: 20.1.5
2. CockroachDB Kubernetes Operator: 1.7.3

## Create a CockroachDB namespace

```bash
oc create namespace cockroachdb
```

```bash
namespace/cockroachdb created
```

## Install the CockroachDB Operator

Navigate to the OperatorHub in the OpenShift web console

Search for `cockroach` and select the tile for CockroachDB. Do not click on the `Marketplace` labeled tile as it requires a subscription.

![crc_cockroach_1](https://user-images.githubusercontent.com/461296/93793269-9b12f100-fc04-11ea-97ba-d6693b30817e.png)

Click the tile, a self-guided dialogue will appear. Click install, another dialogue will appear.

![operator_dialogue](https://user-images.githubusercontent.com/461296/94025412-78a8e100-fd86-11ea-8caa-cb83c1e4bb57.png)

We're going to leave everything as is except for the drop down for namespace. We're going to install the CockroachDB Operator into the namespace we created earlier. Click install.

![crdb_os1](https://user-images.githubusercontent.com/461296/94025061-19e36780-fd86-11ea-9ec2-9a20bae7d917.png)

Once the operator is installed, go to the installed operators section and select CockroachDB. Click create instance.

![crdb_os2](https://user-images.githubusercontent.com/461296/94025059-19e36780-fd86-11ea-98a0-852c2df09910.png)

We're going to leave everything as is. By default, we're going to deploy a 3 node secure cluster. Click create.

![crdb_os3](https://user-images.githubusercontent.com/461296/94025058-19e36780-fd86-11ea-85d1-ce980e48ae57.png)

Now we can navigate to the Pods section and observe how our pods are created.

![crdb_os4](https://user-images.githubusercontent.com/461296/94025057-194ad100-fd86-11ea-9a71-9cae35352504.png)

At this point, you may use the standard `kubectl` commands to inspect your cluster.

```bash
kubectl get pods --namespace=cockroachdb
```

An equivalent OpenShift command is

```bash
oc get pods --namespace=cockroachdb
```

```bash
NAME                                  READY   STATUS    RESTARTS   AGE
cockroach-operator-65c4f6df45-h5r5n   1/1     Running   0          6m11s
crdb-tls-example-0                    1/1     Running   0          3m15s
crdb-tls-example-1                    1/1     Running   0          103s
crdb-tls-example-2                    1/1     Running   0          89s
```

If you prefer to set your namespace preferences instead of typing the namespace each time, feel free to issue the command below.

```bash
oc config set-context --current --namespace=cockroachdb
```

```bash
Context "default/api-crc-testing:6443/kube:admin" modified.
```

kubectl equivalent command is below

```bash
kubectl config set-context --current --namespace=default
```

Validate the namespace

```bash
oc config view --minify | grep namespace:
```

or with kubectl

```bash
kubectl config view --minify | grep namespace:
```

```bash
    namespace: cockroachdb
```

From this point on, I will stick to the OpenShift nomenclature for cluster commands.

## Create a secure client pod to connect to the cluster

Now that we have a CockroachDB cluster running in the OpenShift environment, let's connect to the cluster using a secure client.

My pod YAML for the secure client looks like so

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: crdb-client-secure
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: crdb-tls-example
    app.kubernetes.io/name: cockroachdb
spec:
  serviceAccountName: cockroach-operator-role
  containers:
  - name: crdb-client-secure
    image: registry.connect.redhat.com/cockroachdb/cockroach:v20.1.5
    imagePullPolicy: IfNotPresent
    volumeMounts:
    - name: client-certs
      mountPath: /cockroach/cockroach-certs/
    # Keep a pod open indefinitely so kubectl exec can be used to get a shell to it
    # and run cockroach client commands, such as cockroach sql, cockroach node status, etc.
    command:
    - sleep
    - "2147483648" # 2^31
  # This pod isn't doing anything important, so don't bother waiting to terminate it.
  terminationGracePeriodSeconds: 0
  volumes:
  - name: client-certs
    projected:
        sources:
          - secret:
              name: crdb-tls-example-node
              items:
                - key: ca.crt
                  path: ca.crt
          - secret:
              name: crdb-tls-example-root
              items:
                - key: tls.crt
                  path: client.root.crt
                - key: tls.key
                  path: client.root.key
        defaultMode: 256
```

In the Pods section, click on create pod button and paste the yaml code.

![crdb_os5](https://user-images.githubusercontent.com/461296/94025056-194ad100-fd86-11ea-846c-a16f885af01c.png)

Click create and wait for the pod to run.

Going back to the Pods section, we can see our new client pod running.

![crdb_os6](https://user-images.githubusercontent.com/461296/94025054-194ad100-fd86-11ea-9459-3c78eb8ad397.png)

## Connect to CockroachDB

```bash
oc exec -it crdb-client-secure --namespace cockroachdb -- ./cockroach sql --certs-dir=/cockroach/cockroach-certs/ --host=crdb-tls-example-public
```

```bash
#
# Welcome to the CockroachDB SQL shell.
# All statements must be terminated by a semicolon.
# To exit, type: \q.
#
# Server version: CockroachDB CCL v20.1.5 (x86_64-unknown-linux-gnu, built 2020/08/24 19:52:08, go1.13.9) (same version as client)
# Cluster ID: 0813c343-c86b-4be8-9ad0-477cdb5db749
#
# Enter \? for a brief introduction.
#
root@crdb-tls-example-public:26257/defaultdb> \q
```

And we're in! Let's step through some of the intricacies of what we just did.

We spun up a secure client pod, it belongs to the same namespace and the same lables as our CockroachDB StatefulSet, we are using the same image as the cluster, we're mounting the operator's generated certs. If at this point you're not seeing the same results, it helps to shell into the client pod and see if certs are available.

```bash
oc exec -it crdb-client-secure sh --namespace=cockroachdb
```

```bash
kubectl exec [POD] [COMMAND] is DEPRECATED and will be removed in a future version. Use kubectl kubectl exec [POD] -- [COMMAND] instead.
# cd /cockroach/cockroach-certs
# ls
ca.crt           client.root.key
client.root.crt
```

The other piece that can trip you up is the host to use for the `--host` flag. You can use the public service name.

```bash
oc get services --namespace=cockroachdb
```

```bash
NAME                      TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)              AGE
crdb-tls-example          ClusterIP   None             <none>        26257/TCP,8080/TCP   14m
crdb-tls-example-public   ClusterIP   172.25.180.197   <none>        26257/TCP,8080/TCP   14m
```

## Run a sample workload

We are going to initialize the Movr workload

```bash
oc exec -it crdb-client-secure --namespace cockroachdb -- ./cockroach workload init movr 'postgresql://root@crdb-tls-example-0.crdb-tls-example.cockroachdb:26257?sslcert=%2Fcockroach%2Fcockroach-certs%2Fclient.root.crt&sslkey=%2Fcockroach%2Fcockroach-certs%2Fclient.root.key&sslmode=verify-full&sslrootcert=%2Fcockroach%2Fcockroach-certs%2Fca.crt'
```

* Pro tip: navigating to the CockroachDB logs in the OpenShift console will reveal the jdbc url for Cockroach.

![crdb_logs_openshift](https://user-images.githubusercontent.com/461296/94047412-0e9d3580-fda0-11ea-8965-17d8c49d0248.png)

Running a Movr workload is as follows:

```bash
oc exec -it crdb-client-secure --namespace cockroachdb -- ./cockroach workload run movr --duration=3m --tolerate-errors --max-rate=20 --concurrency=1 --display-every=10s 'postgresql://root@crdb-tls-example-0.crdb-tls-example.cockroachdb:26257?sslcert=%2Fcockroach%2Fcockroach-certs%2Fclient.root.crt&sslkey=%2Fcockroach%2Fcockroach-certs%2Fclient.root.key&sslmode=verify-full&sslrootcert=%2Fcockroach%2Fcockroach-certs%2Fca.crt'
```

## Demonstrate resilience

While the workload is running, kill a node and see how the cluster heals itself.

![kill_node](https://user-images.githubusercontent.com/461296/94047411-0e049f00-fda0-11ea-91e4-fa11034a3da2.png)

![confirm](https://user-images.githubusercontent.com/461296/94047410-0e049f00-fda0-11ea-8c50-ecf6e3de2f4f.png)

![running](https://user-images.githubusercontent.com/461296/94047409-0e049f00-fda0-11ea-8d3c-6835a1e060b7.png)

Notice the start time

We can also kill a node using CLI

```bash
oc delete pod crdb-tls-example-1 --namespace=cockroachdb
```

```bash
pod "crdb-tls-example-1" deleted
```

Let's also look at the CockroachDB Admin UI for the health status. Port forward the Admin UI in OpenShift

```bash
oc port-forward --namespace=cockroachdb crdb-tls-example-0 8080 &
```

```bash
Forwarding from [::1]:8080 -> 8080
```

Create a SQL user to access the Web UI

```bash
oc exec -it crdb-client-secure --namespace cockroachdb -- ./cockroach sql --certs-dir=/cockroach/cockroach-certs/ --host=crdb-tls-example-public
```

```sql
CREATE USER craig WITH PASSWORD 'cockroach';
```

Navigate to the [Admin UI](https://localhost:8080/#/overview/list) and enter user `craig` and password `cockroach` to login.

Notice the replication status

![kill_node_admin_ui](https://user-images.githubusercontent.com/461296/94047408-0e049f00-fda0-11ea-8616-af051c0d2523.png)

After the cluster heals itself, check the replication status again.

![healed](https://user-images.githubusercontent.com/461296/94047407-0d6c0880-fda0-11ea-9473-5699a945d571.png)
