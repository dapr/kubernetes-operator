package controlplane

import (
	"context"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/internal/controller/operator/instance"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/client"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/gc"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/helm"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

func NewStatusAction(l logr.Logger) Action {
	return &StatusAction{
		engine:        helm.NewEngine(),
		l:             l.WithName("action").WithName("status"),
		subscriptions: make(map[string]struct{}),
		gc:            gc.New(),
	}
}

// StatusAction computes the state of a DaprControlPlane resource out of the owned DaprInstance resource.
type StatusAction struct {
	engine        *helm.Engine
	gc            *gc.GC
	l             logr.Logger
	subscriptions map[string]struct{}
}

func (a *StatusAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	return b, nil
}

func (a *StatusAction) Run(ctx context.Context, rr *ReconciliationRequest) error {
	di, err := rr.Dapr.OperatorV1alpha1().DaprInstances(rr.Resource.Namespace).Get(
		ctx,
		instance.DaprInstanceResourceName,
		metav1.GetOptions{},
	)

	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	for i := range di.Status.Conditions {
		meta.SetStatusCondition(&rr.Resource.Status.Conditions, di.Status.Conditions[i])
	}

	rr.Resource.Status.Chart = di.Status.Chart

	return nil
}

func (a *StatusAction) Cleanup(_ context.Context, _ *ReconciliationRequest) error {
	return nil
}
