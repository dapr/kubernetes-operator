package instance

import (
	"context"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/client"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/helm"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

func NewChartAction(l logr.Logger) Action {
	return &ChartAction{
		engine: helm.NewEngine(),
		l:      l.WithName("action").WithName("chart"),
	}
}

type ChartAction struct {
	engine *helm.Engine
	l      logr.Logger
}

func (a *ChartAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	return b, nil
}

func (a *ChartAction) Run(_ context.Context, rc *ReconciliationRequest) error {
	if rc.Resource.Status.Chart == nil {
		rc.Resource.Status.Chart = &v1alpha1.ChartMeta{}
	}

	// TODO: maybe cache the chart
	if rc.Resource.Spec.Chart != nil {
		c, err := a.engine.Load(rc.Resource.Spec.Chart.Repo, rc.Resource.Spec.Chart.Name, rc.Resource.Spec.Chart.Version)
		if err != nil {
			return err
		}

		rc.Chart = c

		rc.Resource.Status.Chart.Repo = rc.Resource.Spec.Chart.Repo
	} else {
		rc.Resource.Status.Chart.Repo = "embedded"
	}

	rc.Resource.Status.Chart.Version = rc.Chart.Metadata.Version
	rc.Resource.Status.Chart.Name = rc.Chart.Metadata.Name

	return nil
}

func (a *ChartAction) Cleanup(_ context.Context, _ *ReconciliationRequest) error {
	return nil
}
