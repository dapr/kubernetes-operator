package instance

import (
	"context"

	daprApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/client"
	"helm.sh/helm/v3/pkg/chart"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

const (
	DaprInstanceFinalizerName = "instance.operator.dapr.io/finalizer"
	DaprInstanceResourceName  = "dapr-instance"
)

type ReconciliationRequest struct {
	*client.Client
	types.NamespacedName

	Reconciler  *Reconciler
	ClusterType controller.ClusterType
	Resource    *daprApi.DaprInstance
	Chart       *chart.Chart
	Overrides   map[string]interface{}
}

type Action interface {
	Configure(ctx context.Context, c *client.Client, b *builder.Builder) (*builder.Builder, error)
	Run(ctx context.Context, rc *ReconciliationRequest) error
	Cleanup(ctx context.Context, rc *ReconciliationRequest) error
}
