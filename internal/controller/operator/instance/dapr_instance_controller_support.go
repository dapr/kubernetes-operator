package instance

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dapr/kubernetes-operator/pkg/helm"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	ctrlCli "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/dapr/kubernetes-operator/pkg/controller/predicates"
)

func gcSelector(ctx context.Context, rc *ReconciliationRequest) (labels.Selector, error) {
	c, err := rc.Chart(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load chart: %w", err)
	}

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
		selection.Exists,
		[]string{})
	if err != nil {
		return nil, fmt.Errorf("cannot determine generation requirement: %w", err)
	}

	version, err := labels.NewRequirement(
		helm.ReleaseVersion,
		selection.Equals,
		[]string{c.Version()})
	if err != nil {
		return nil, fmt.Errorf("cannot determine release version requirement: %w", err)
	}

	selector := labels.NewSelector().
		Add(*namespace).
		Add(*name).
		Add(*generation).
		Add(*version)

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

func dependantWithLabels(opts ...predicates.DependentPredicateOption) predicate.Predicate {
	dp := &predicates.DependentPredicate{}

	for i := range opts {
		dp = opts[i](dp)
	}

	return predicate.And(
		&predicates.HasLabel{
			Name: helm.ReleaseName,
		},
		&predicates.HasLabel{
			Name: helm.ReleaseNamespace,
		},
		dp,
	)
}

func currentReleaseSelector(ctx context.Context, rc *ReconciliationRequest) (labels.Selector, error) {
	c, err := rc.Chart(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load chart: %w", err)
	}

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

	version, err := labels.NewRequirement(
		helm.ReleaseVersion,
		selection.Equals,
		[]string{c.Version()})
	if err != nil {
		return nil, fmt.Errorf("cannot determine release version requirement: %w", err)
	}

	selector := labels.NewSelector().
		Add(*namespace).
		Add(*name).
		Add(*generation).
		Add(*version)

	return selector, nil
}
