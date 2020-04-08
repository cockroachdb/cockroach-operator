package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
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
