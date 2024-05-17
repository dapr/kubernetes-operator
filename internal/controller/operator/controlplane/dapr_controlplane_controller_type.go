package controlplane

import (
	"context"

	daprApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/client"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

const (
	DaprControlPlaneFinalizerName = "controlplane.operator.dapr.io/finalizer"
	DaprControlPlaneResourceName  = "dapr-control-plane"
)

type ReconciliationRequest struct {
	*client.Client
	types.NamespacedName

	Reconciler  *Reconciler
	ClusterType controller.ClusterType
	Resource    *daprApi.DaprControlPlane
}

type Action interface {
	Configure(ctx context.Context, c *client.Client, b *builder.Builder) (*builder.Builder, error)
	Run(ctx context.Context, rc *ReconciliationRequest) error
	Cleanup(ctx context.Context, rc *ReconciliationRequest) error
}
