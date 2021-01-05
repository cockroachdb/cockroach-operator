/*
Copyright 2021 The Cockroach Authors

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

package controller

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

func requeueIfError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

func noRequeue() (ctrl.Result, error) {
	return requeueIfError(nil)
}

func requeueAfter(interval time.Duration, err error) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: interval}, err
}

func requeueImmediately() (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, nil
}
