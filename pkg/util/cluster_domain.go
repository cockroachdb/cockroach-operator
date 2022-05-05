/*
Copyright 2022 The Cockroach Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"context"
	"net"
	"sync"
	"time"
)

var (
	once          = &sync.Once{}
	clusterDomain = "cluster.local"
)

// GetClusterDomain returns the cluster domain of the k8s cluster.
// It is auto-detected by lazily running a DNS query.
// It defaults to "cluster.local" if we cannot determine the domain.
func GetClusterDomain() string {
	once.Do(func() {
		// We try to lookup a non-FQDN that *should* always exist in the
		// k8s's domain.
		// Reference: https://stackoverflow.com/a/59162874
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		const host = "kubernetes.default.svc"
		cname, err := net.DefaultResolver.LookupCNAME(ctx, host)
		if err == nil {
			clusterDomain = cname[len(host)+1 : len(cname)-1]
		}
	})
	return clusterDomain
}
