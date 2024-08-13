package instance

import (
	"context"
	"fmt"

	"github.com/dapr/kubernetes-operator/api/operator/v1alpha1"
	"github.com/dapr/kubernetes-operator/pkg/controller/client"
	"github.com/go-logr/logr"
	helme "github.com/lburgazzoli/k8s-manifests-renderer-helm/engine"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

const (
	ChartRepoUsernameKey = "username"
	ChartRepoPasswordKey = "password"
	ChartRepoEmbedded    = "embedded"
)

func NewChartAction(l logr.Logger) Action {
	return &ChartAction{
		engine: helme.New(),
		l:      l.WithName("action").WithName("chart"),
	}
}

type ChartAction struct {
	engine *helme.Instance
	l      logr.Logger
}

func (a *ChartAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	return b, nil
}

func (a *ChartAction) Run(ctx context.Context, rc *ReconciliationRequest) error {
	c, err := rc.Chart(ctx)
	if err != nil {
		return fmt.Errorf("cannot load chart: %w", err)
	}

	if rc.Resource.Status.Chart == nil {
		rc.Resource.Status.Chart = &v1alpha1.ChartMeta{}
	}

	rc.Resource.Status.Chart.Repo = ChartRepoEmbedded
	rc.Resource.Status.Chart.Version = c.Version()
	rc.Resource.Status.Chart.Name = c.Name()

	if rc.Resource.Spec.Chart != nil {
		rc.Resource.Status.Chart.Repo = rc.Resource.Spec.Chart.Repo
	}

	return nil
}

func (a *ChartAction) Cleanup(_ context.Context, _ *ReconciliationRequest) error {
	return nil
}
