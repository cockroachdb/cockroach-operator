# Developer Notes

## Testing db

kubectl exec into a pod

```
cockroach sql
CREATE TABLE bank.accounts (id INT PRIMARY KEY, balance DECIMAL);
INSERT INTO bank.accounts VALUES (1, 1000.50);
SELECT * FROM bank.accounts;
```

## etcd in WORKSPACE file

These are the commands used to upload the etcd binary to a gs bucket.
```
curl -LO curl -LO https://storage.googleapis.com/etcd/v3.3.11/etcd-v3.3.11-linux-amd64.tar.gz
tar xzf etcd-v3.3.11-linux-amd64.tar.gz
history| grep gsutil
cd etcd-v3.3.11-linux-amd64
gsutil cp etcd gs://crdb-bazel-artifacts
sha256sum etcd
```
