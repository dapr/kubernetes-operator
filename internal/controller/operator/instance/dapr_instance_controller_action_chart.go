package instance

import (
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/chart"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/client"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/helm"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

const (
	ChartRepoUsernameKey = "username"
	ChartRepoPasswordKey = "password"
	ChartRepoEmbedded    = "embedded"
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

func (a *ChartAction) Run(ctx context.Context, rc *ReconciliationRequest) error {
	if rc.Resource.Status.Chart == nil {
		rc.Resource.Status.Chart = &v1alpha1.ChartMeta{}
	}

	rc.Resource.Status.Chart.Repo = ChartRepoEmbedded

	// TODO: maybe cache the chart
	if rc.Resource.Spec.Chart != nil {
		c, err := a.loadChart(ctx, rc)
		if err != nil {
			return err
		}

		rc.Chart = c

		rc.Resource.Status.Chart.Repo = rc.Resource.Spec.Chart.Repo
	}

	rc.Resource.Status.Chart.Version = rc.Chart.Metadata.Version
	rc.Resource.Status.Chart.Name = rc.Chart.Metadata.Name

	return nil
}

func (a *ChartAction) Cleanup(_ context.Context, _ *ReconciliationRequest) error {
	return nil
}

func (a *ChartAction) loadChart(ctx context.Context, rc *ReconciliationRequest) (*chart.Chart, error) {
	opts := helm.ChartOptions{}
	opts.Name = rc.Resource.Spec.Chart.Name
	opts.RepoURL = rc.Resource.Spec.Chart.Repo
	opts.Version = rc.Resource.Spec.Chart.Version

	if rc.Resource.Spec.Chart.Secret != "" {
		s, err := rc.Client.CoreV1().Secrets(rc.Resource.Namespace).Get(
			ctx,
			rc.Resource.Spec.Chart.Secret,
			metav1.GetOptions{},
		)

		if err != nil {
			return nil, fmt.Errorf("unable to fetch secret %s, %w", rc.Resource.Spec.Chart.Secret, err)
		}

		if v, ok := s.Data[ChartRepoUsernameKey]; ok {
			opts.Username = string(v)
		}
		if v, ok := s.Data[ChartRepoPasswordKey]; ok {
			opts.Password = string(v)
		}
	}

	c, err := a.engine.Load(opts)
	if err != nil {
		return nil, err
	}

	return c, nil
}
