#!/usr/bin/env bash

# Copyright 2021 The Cockroach Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

# Script used to build and install the operator on k8s
# The script then deploys a cr and creates a roach db

make k8s/apply
cat <<EOF | kubectl apply -f -
apiVersion: crdb.cockroachlabs.com/v1alpha1
kind: CrdbCluster
metadata:
  name: crdb-update
spec:
  dataStore:
    pvc:
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 5Gi
        volumeMode: Filesystem
  tlsEnabled: true
  image:
    name: cockroachdb/cockroach:v19.2.6
  nodes: 3
EOF
sleep 10
kubectl rollout status -w statefulset/crdb-update
sleep 10
kubectl edit crdbclusters.crdb.cockroachlabs.com crdb-update
kubectl get po --watch
