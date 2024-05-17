package controlplane

import (
	"context"
	"fmt"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/resources"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/internal/controller/operator/instance"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	daprApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	daprAc "github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/client/applyconfiguration/operator/v1alpha1"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/client"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/predicates"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func NewApplyAction(l logr.Logger) Action {
	return &ApplyAction{
		l:             l.WithName("action").WithName("apply"),
		subscriptions: make(map[string]struct{}),
	}
}

type ApplyAction struct {
	l             logr.Logger
	subscriptions map[string]struct{}
}

func (a *ApplyAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	b = b.Owns(&daprApi.DaprInstance{}, builder.WithPredicates(
		predicate.Or(
			predicate.GenerationChangedPredicate{},
			predicates.StatusChanged{},
		)))

	return b, nil
}

func (a *ApplyAction) Run(ctx context.Context, rr *ReconciliationRequest) error {
	spec := daprAc.DaprInstanceSpec()

	if rr.Resource.Spec.Values != nil {
		values := daprAc.JSON()
		values.RawMessage = rr.Resource.Spec.Values.RawMessage

		spec = spec.WithValues(values)
	}

	_, err := rr.Dapr.OperatorV1alpha1().DaprInstances(rr.Resource.Namespace).Apply(
		ctx,
		daprAc.DaprInstance(instance.DaprInstanceResourceName, rr.Resource.Namespace).
			WithOwnerReferences(resources.WithOwnerReference(rr.Resource)).
			WithSpec(spec),
		metav1.ApplyOptions{
			FieldManager: controller.FieldManager,
		})

	if err != nil {
		return fmt.Errorf("failure to apply changes to %s: %w", rr.NamespacedName, err)
	}

	return nil
}

func (a *ApplyAction) Cleanup(_ context.Context, _ *ReconciliationRequest) error {
	return nil
}
