package instance

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/dapr/kubernetes-operator/pkg/conditions"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/dapr/kubernetes-operator/pkg/controller/client"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

func NewConditionsAction(l logr.Logger) Action {
	return &ConditionsAction{
		l: l.WithName("action").WithName("conditions"),
	}
}

type ConditionsAction struct {
	l logr.Logger
}

func (a *ConditionsAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	return b, nil
}

func (a *ConditionsAction) Run(ctx context.Context, rc *ReconciliationRequest) error {
	crs, err := currentReleaseSelector(ctx, rc)
	if err != nil {
		return fmt.Errorf("cannot compute current release selector: %w", err)
	}

	deployments, readyDeployments, err := a.deployments(ctx, rc, crs)
	if err != nil {
		return fmt.Errorf("cannot count deployments: %w", err)
	}

	statefulSets, readyReplicaSets, err := a.statefulSets(ctx, rc, crs)
	if err != nil {
		return fmt.Errorf("cannot count stateful sets: %w", err)
	}

	var readyCondition metav1.Condition

	if deployments+statefulSets > 0 {
		reason := conditions.ReasonReady
		status := metav1.ConditionTrue

		if readyDeployments+readyReplicaSets != deployments+statefulSets {
			reason = conditions.ReasonInProgress
			status = metav1.ConditionFalse
		}

		readyCondition = metav1.Condition{
			Type:               conditions.TypeReady,
			Status:             status,
			Reason:             reason,
			ObservedGeneration: rc.Resource.Generation,
			Message: fmt.Sprintf("%d/%d deployments ready, statefulSets ready %d/%d",
				readyDeployments, deployments,
				readyReplicaSets, statefulSets),
		}
	} else {
		readyCondition = metav1.Condition{
			Type:               conditions.TypeReady,
			Status:             metav1.ConditionFalse,
			Reason:             conditions.ReasonInProgress,
			Message:            "no deployments/replicasets",
			ObservedGeneration: rc.Resource.Generation,
		}
	}

	meta.SetStatusCondition(&rc.Resource.Status.Conditions, readyCondition)

	return nil
}

func (a *ConditionsAction) Cleanup(_ context.Context, _ *ReconciliationRequest) error {
	return nil
}

func (a *ConditionsAction) deployments(ctx context.Context, rc *ReconciliationRequest, selector labels.Selector) (int, int, error) {
	objects, err := rc.Client.AppsV1().Deployments(rc.Resource.Spec.Deployment.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return 0, 0, fmt.Errorf("cannot list deployments: %w", err)
	}

	ready := 0

	for i := range objects.Items {
		if conditions.ConditionStatus(objects.Items[i], appsv1.DeploymentAvailable) == corev1.ConditionTrue {
			ready++
		}
	}

	return len(objects.Items), ready, nil
}

func (a *ConditionsAction) statefulSets(ctx context.Context, rc *ReconciliationRequest, selector labels.Selector) (int, int, error) {
	objects, err := rc.Client.AppsV1().StatefulSets(rc.Resource.Spec.Deployment.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return 0, 0, fmt.Errorf("cannot list stateful sets: %w", err)
	}

	ready := 0

	for i := range objects.Items {
		if objects.Items[i].Status.Replicas == 0 {
			continue
		}

		if objects.Items[i].Status.Replicas == objects.Items[i].Status.ReadyReplicas {
			ready++
		}
	}

	return len(objects.Items), ready, nil
}
