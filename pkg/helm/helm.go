package helm

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	ReleaseGeneration = "helm.operator.dapr.io/release.generation"
	ReleaseName       = "helm.operator.dapr.io/release.name"
	ReleaseNamespace  = "helm.operator.dapr.io/release.namespace"

	ChartsDir = "helm-charts/dapr"
)

type Options struct {
	ChartsDir string
}

//nolint:wrapcheck
func ReleaseSelector() (labels.Selector, error) {
	hasReleaseNameLabel, err := labels.NewRequirement(ReleaseName, selection.Exists, []string{})
	if err != nil {
		return nil, err
	}

	hasReleaseNamespaceLabel, err := labels.NewRequirement(ReleaseNamespace, selection.Exists, []string{})
	if err != nil {
		return nil, err
	}

	selector := labels.NewSelector().
		Add(*hasReleaseNameLabel).
		Add(*hasReleaseNamespaceLabel)

	return selector, nil
}
