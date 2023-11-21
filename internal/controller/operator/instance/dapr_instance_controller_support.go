package instance

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/helm"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	ctrlCli "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/predicates"
)

func gcSelector(rc *ReconciliationRequest) (labels.Selector, error) {

	namespace, err := labels.NewRequirement(
		helm.ReleaseNamespace,
		selection.Equals,
		[]string{rc.Resource.Namespace})

	if err != nil {
		return nil, fmt.Errorf("cannot determine release namespace requirement: %w", err)
	}

	name, err := labels.NewRequirement(
		helm.ReleaseName,
		selection.Equals,
		[]string{rc.Resource.Name})

	if err != nil {
		return nil, fmt.Errorf("cannot determine release name requirement: %w", err)
	}

	generation, err := labels.NewRequirement(
		helm.ReleaseGeneration,
		selection.LessThan,
		[]string{strconv.FormatInt(rc.Resource.Generation, 10)})

	if err != nil {
		return nil, fmt.Errorf("cannot determine generation requirement: %w", err)
	}

	selector := labels.NewSelector().
		Add(*namespace).
		Add(*name).
		Add(*generation)

	return selector, nil
}

func labelsToRequest(_ context.Context, object ctrlCli.Object) []reconcile.Request {
	allLabels := object.GetLabels()
	if allLabels == nil {
		return nil
	}
	name := allLabels[helm.ReleaseName]
	if name == "" {
		return nil
	}
	namespace := allLabels[helm.ReleaseNamespace]
	if namespace == "" {
		return nil
	}

	return []reconcile.Request{{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}}
}

func dependantWithLabels(watchUpdate bool, watchDelete bool, watchStatus bool) predicate.Predicate {
	return predicate.And(
		&predicates.HasLabel{
			Name: helm.ReleaseName,
		},
		&predicates.HasLabel{
			Name: helm.ReleaseNamespace,
		},
		&predicates.DependentPredicate{
			WatchUpdate: watchUpdate,
			WatchDelete: watchDelete,
			WatchStatus: watchStatus,
		},
	)
}

func CurrentReleaseSelector(rc *ReconciliationRequest) (labels.Selector, error) {
	namespace, err := labels.NewRequirement(
		helm.ReleaseNamespace,
		selection.Equals,
		[]string{rc.Resource.Namespace})

	if err != nil {
		return nil, fmt.Errorf("cannot determine release namespace requirement: %w", err)
	}

	name, err := labels.NewRequirement(
		helm.ReleaseName,
		selection.Equals,
		[]string{rc.Resource.Name})

	if err != nil {
		return nil, fmt.Errorf("cannot determine release name requirement: %w", err)
	}

	generation, err := labels.NewRequirement(
		helm.ReleaseGeneration,
		selection.Equals,
		[]string{strconv.FormatInt(rc.Resource.Generation, 10)})

	if err != nil {
		return nil, fmt.Errorf("cannot determine generation requirement: %w", err)
	}

	selector := labels.NewSelector().
		Add(*namespace).
		Add(*name).
		Add(*generation)

	return selector, nil
}
