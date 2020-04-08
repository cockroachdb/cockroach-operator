package actor

import (
	"context"
	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/resource"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type NotReadyErr struct {
	Err error
}

func (e NotReadyErr) Error() string {
	return e.Err.Error()
}

type Actor interface {
	Handles([]api.ClusterCondition) bool
	Act(context.Context, *resource.Cluster) error
}

func NewOperatorActions(scheme *runtime.Scheme, cl client.Client, config *rest.Config) []Actor {
	return []Actor{
		newDeploy(scheme, cl),
		newInitialize(scheme, cl, config),
	}
}

var Log = logf.Log.WithName("action")

func newAction(atype string, scheme *runtime.Scheme, cl client.Client) action {
	return action{
		log:    Log.WithValues("action", atype),
		client: cl,
		scheme: scheme,
	}
}

type action struct {
	log    logr.Logger
	client client.Client
	scheme *runtime.Scheme
}
