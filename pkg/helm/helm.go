package helm

import (
	daprApi "github.com/dapr/kubernetes-operator/api/operator/v1beta1"
	helme "github.com/lburgazzoli/k8s-manifests-renderer-helm/engine"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	ReleaseGeneration = "helm.operator.dapr.io/release.generation"
	ReleaseName       = "helm.operator.dapr.io/release.name"
	ReleaseVersion    = "helm.operator.dapr.io/release.version"

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

	selector := labels.NewSelector().
		Add(*hasReleaseNameLabel)

	return selector, nil
}

func IsSameChart(c *helme.Chart, meta *daprApi.ChartMeta) bool {
	if c == nil || meta == nil {
		return false
	}

	return c.Name() == meta.Name &&
		c.Version() == meta.Version &&
		c.Repo() == meta.Repo
}
