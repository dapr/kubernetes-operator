package helm

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

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
