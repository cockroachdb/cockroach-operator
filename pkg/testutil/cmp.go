package testutil

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var RuntimeObjCmpOpts = []cmp.Option{
	cmpopts.IgnoreTypes(metav1.TypeMeta{}),
	cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ResourceVersion", "SelfLink", "Generation", "CreationTimestamp"),
}
