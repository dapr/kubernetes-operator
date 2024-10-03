package instance

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dapr/kubernetes-operator/pkg/helm"

	"github.com/dapr/kubernetes-operator/pkg/controller/predicates"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func gcSelector(ctx context.Context, rc *ReconciliationRequest) (labels.Selector, error) {
	c, err := rc.Chart(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load chart: %w", err)
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
		Add(*name).
		Add(*generation).
		Add(*version)

	return selector, nil
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
		dp,
	)
}
func partialDependantWithLabels(opts ...predicates.PartialDependentPredicateOption) predicate.Predicate {
	dp := &predicates.PartialDependentPredicate{}

	for i := range opts {
		dp = opts[i](dp)
	}

	return predicate.And(
		&predicates.HasLabel{
			Name: helm.ReleaseName,
		},
		dp,
	)
}

func currentReleaseSelector(ctx context.Context, rc *ReconciliationRequest) (labels.Selector, error) {
	c, err := rc.Chart(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load chart: %w", err)
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
		Add(*name).
		Add(*generation).
		Add(*version)

	return selector, nil
}
