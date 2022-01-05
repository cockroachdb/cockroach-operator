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
	"fmt"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

// CheckIfAPIVersionKindAvailable checks that if the given kind is present in the given API version
func CheckIfAPIVersionKindAvailable(config *rest.Config, groupVersion, kind string) (found bool) {
	discoveryClient, err := GetDiscoveryClient(config)
	if err != nil {
		_ = fmt.Errorf("error while getting the discovery client for the rest config: %s", err.Error())
		return
	}

	resources, err := discoveryClient.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		if apierrs.IsNotFound(err) {
			fmt.Printf("%s not found", groupVersion)
			return
		}
		return found
	}

	// if kind is empty then return GroupVersion availability
	if kind == "" {
		found = true
	} else {
		for i := range resources.APIResources {
			if resources.APIResources[i].Kind == kind {
				found = true
				break
			}
		}
	}
	return
}

// GetDiscoveryClient initializes the discovery client using the rest config.
func GetDiscoveryClient(restConfig *rest.Config) (*discovery.DiscoveryClient, error) {
	discClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return discClient, nil
}
