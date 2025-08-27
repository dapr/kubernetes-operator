package instance

import (
	"context"
	"fmt"

	"github.com/lburgazzoli/k8s-manifests-renderer-helm/engine/customizers/values"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	helme "github.com/lburgazzoli/k8s-manifests-renderer-helm/engine"

	daprApi "github.com/dapr/kubernetes-operator/api/operator/v1alpha1"
	"github.com/dapr/kubernetes-operator/pkg/controller"
	"github.com/dapr/kubernetes-operator/pkg/controller/client"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

const (
	DaprInstanceFinalizerName = "instance.operator.dapr.io/finalizer"
	DaprInstanceResourceName  = "dapr-instance"
)

type ReconciliationRequest struct {
	*client.Client
	types.NamespacedName

	Reconciler  *Reconciler
	ClusterType controller.ClusterType
	Resource    *daprApi.DaprInstance
	Helm        Helm
}

type Helm struct {
	engine         *helme.Instance
	chart          *helme.Chart
	chartDir       string
	ChartValues    map[string]interface{}
	chartOverrides map[string]interface{}
}

func (rr *ReconciliationRequest) Chart(ctx context.Context) (*helme.Chart, error) {
	if rr.Helm.chart != nil {
		return rr.Helm.chart, nil
	}

	cs := helme.ChartSpec{
		Name: rr.Helm.chartDir,
	}

	if rr.Resource.Spec.Chart != nil {
		cs.Name = rr.Resource.Spec.Chart.Name
		cs.Repo = rr.Resource.Spec.Chart.Repo
		cs.Version = rr.Resource.Spec.Chart.Version
	}

	chartOpts, err := rr.computeChartOptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to compute chart opetions: %w", err)
	}

	c, err := rr.Helm.engine.Load(
		ctx,
		cs,
		chartOpts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failure loading chart: %w", err)
	}

	rr.Helm.chart = c

	return rr.Helm.chart, nil
}

func (rr *ReconciliationRequest) computeChartOptions(ctx context.Context) ([]helme.ChartOption, error) {
	chartOpts := make([]helme.ChartOption, 0)
	chartOpts = append(chartOpts, helme.WithOverrides(rr.Helm.chartOverrides))
	chartOpts = append(chartOpts, helme.WithValuesCustomizers(values.JQ(autoPullPolicySidecarInjector)))

	if rr.Resource.Spec.Chart != nil && rr.Resource.Spec.Chart.Secret != "" {
		s, err := rr.Client.CoreV1().Secrets(rr.Resource.Namespace).Get(
			ctx,
			rr.Resource.Spec.Chart.Secret,
			metav1.GetOptions{},
		)

		switch {
		case k8serrors.IsNotFound(err):
			break
		case err != nil:
			return nil, fmt.Errorf("unable to fetch secret %s, %w", rr.Resource.Spec.Chart.Secret, err)
		default:
			if v, ok := s.Data[ChartRepoUsernameKey]; ok {
				chartOpts = append(chartOpts, helme.WithUsername(string(v)))
			}

			if v, ok := s.Data[ChartRepoPasswordKey]; ok {
				chartOpts = append(chartOpts, helme.WithPassword(string(v)))
			}
		}
	}

	return chartOpts, nil
}

type Action interface {
	Configure(ctx context.Context, c *client.Client, b *builder.Builder) (*builder.Builder, error)
	Run(ctx context.Context, rc *ReconciliationRequest) error
	Cleanup(ctx context.Context, rc *ReconciliationRequest) error
}
