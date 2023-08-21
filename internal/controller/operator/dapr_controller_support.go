package operator

import (
	"context"
	"strconv"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/predicates"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	ctrlCli "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func gcSelector(rc *ReconciliationRequest) (labels.Selector, error) {

	namespace, err := labels.NewRequirement(
		DaprReleaseNamespace,
		selection.Equals,
		[]string{rc.Resource.Namespace})

	if err != nil {
		return nil, errors.Wrap(err, "cannot determine release namespace requirement")
	}

	name, err := labels.NewRequirement(
		DaprReleaseName,
		selection.Equals,
		[]string{rc.Resource.Name})

	if err != nil {
		return nil, errors.Wrap(err, "cannot determine release name requirement")
	}

	generation, err := labels.NewRequirement(
		DaprReleaseGeneration,
		selection.LessThan,
		[]string{strconv.FormatInt(rc.Resource.Generation, 10)})

	if err != nil {
		return nil, errors.Wrap(err, "cannot determine generation requirement")
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
	name := allLabels[DaprReleaseName]
	if name == "" {
		return nil
	}
	namespace := allLabels[DaprReleaseNamespace]
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

func dependantWithLabels(watchUpdate bool, watchDelete bool) predicate.Predicate {
	return predicate.And(
		&predicates.HasLabel{
			Name: DaprReleaseName,
		},
		&predicates.HasLabel{
			Name: DaprReleaseNamespace,
		},
		&predicates.DependentPredicate{
			WatchUpdate: watchUpdate,
			WatchDelete: watchDelete,
		},
	)
}
