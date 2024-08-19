package instance

import (
	"context"
	"fmt"

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

	// Deployments

	deployments, err := rc.Client.AppsV1().Deployments(rc.Resource.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: crs.String(),
	})

	if err != nil {
		return fmt.Errorf("cannot list deployments: %w", err)
	}

	readyDeployments := 0

	for i := range deployments.Items {
		if conditions.ConditionStatus(deployments.Items[i], appsv1.DeploymentAvailable) == corev1.ConditionTrue {
			readyDeployments++
		}
	}

	// StatefulSets

	statefulSets, err := rc.Client.AppsV1().StatefulSets(rc.Resource.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: crs.String(),
	})

	if err != nil {
		return fmt.Errorf("cannot list stateful sets: %w", err)
	}

	readyReplicaSets := 0

	for i := range statefulSets.Items {
		if statefulSets.Items[i].Status.Replicas == 0 {
			continue
		}

		if statefulSets.Items[i].Status.Replicas == statefulSets.Items[i].Status.ReadyReplicas {
			readyReplicaSets++
		}
	}

	var readyCondition metav1.Condition

	if len(deployments.Items)+len(statefulSets.Items) > 0 {
		if readyDeployments+readyReplicaSets == len(deployments.Items)+len(statefulSets.Items) {
			readyCondition = metav1.Condition{
				Type:               conditions.TypeReady,
				Status:             metav1.ConditionTrue,
				Reason:             "Ready",
				ObservedGeneration: rc.Resource.Generation,
				Message: fmt.Sprintf("%d/%d deployments ready, statefulSets ready %d/%d",
					readyDeployments, len(deployments.Items),
					readyReplicaSets, len(statefulSets.Items)),
			}
		} else {
			readyCondition = metav1.Condition{
				Type:               conditions.TypeReady,
				Status:             metav1.ConditionFalse,
				Reason:             "InProgress",
				ObservedGeneration: rc.Resource.Generation,
				Message: fmt.Sprintf("%d/%d deployments ready, statefulSets ready %d/%d",
					readyDeployments, len(deployments.Items),
					readyReplicaSets, len(statefulSets.Items)),
			}
		}
	} else {
		readyCondition = metav1.Condition{
			Type:               conditions.TypeReady,
			Status:             metav1.ConditionFalse,
			Reason:             "InProgress",
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
