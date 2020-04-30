package kube

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var IsNotFound = apierrors.IsNotFound

var IgnoreNotFound = client.IgnoreNotFound
