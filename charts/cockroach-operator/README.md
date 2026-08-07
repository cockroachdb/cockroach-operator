# CockroachDB operator Helm chart

This chart is the Helm equivalent of the Kubernetes installation assembled by
`config/default`. It installs the CRD, operator RBAC, manager deployment, and
admission webhooks. `CrdbCluster` objects are intentionally not installed by
the chart.

## Generate and validate

Install Helm, then run:

```text
make helm/chart
```

The target copies the canonical CRD from `config/crd/bases` into `crds/` and
runs `helm lint`. The CRD is kept out of `templates/` because Helm handles CRD
installation and upgrades separately.

## Install

```text
helm upgrade --install cockroach-operator \
  ./charts/cockroach-operator \
  --namespace cockroach-operator-system \
  --create-namespace
```

The operator manages webhook certificates at startup, matching the existing
Kustomize deployment. Set `skipWebhookConfig=true` only when an external
platform manages webhook TLS and admission configuration.