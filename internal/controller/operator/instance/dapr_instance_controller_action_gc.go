package instance

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dapr/kubernetes-operator/pkg/controller"

	"github.com/dapr/kubernetes-operator/pkg/controller/gc"
	"github.com/dapr/kubernetes-operator/pkg/helm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/dapr/kubernetes-operator/pkg/controller/client"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

func NewGCAction(l logr.Logger) Action {
	return &GCAction{
		l:  l.WithName("action").WithName("gc"),
		gc: gc.New(),
	}
}

// GCAction cleanup leftover release resources.
//
// If the HelmInstance spec changes, all the resources get re-rendered which means some of
// them may become obsolete (i.e. if some resources are moved from cluster to namespace
// scope) hence a sort of "garbage collector task" must be executed.
//
// The logic of the task it to delete all the resources that have a generation older than
// current CR one or rendered out of a release version different from the current one. The
// related values are propagated by the controller to all the rendered resources in as a
// set of labels (
//
// - helm.operator.dapr.io/release.generation
// - helm.operator.dapr.io/release.version
//
// The action MUST be executed as the latest action in the reconciliation loop.
type GCAction struct {
	l  logr.Logger
	gc *gc.GC
}

func (a *GCAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	return b, nil
}

func (a *GCAction) Run(ctx context.Context, rc *ReconciliationRequest) error {
	c, err := rc.Chart(ctx)
	if err != nil {
		return fmt.Errorf("cannot load chart: %w", err)
	}

	s, err := gcSelector(ctx, rc)
	if err != nil {
		return fmt.Errorf("cannot compute gc selector: %w", err)
	}

	deleted, err := a.gc.Run(ctx, rc.Client, controller.OperatorNamespace(), s, func(ctx context.Context, obj unstructured.Unstructured) (bool, error) {
		if obj.GetLabels() == nil {
			return false, nil
		}

		gen := obj.GetLabels()[helm.ReleaseGeneration]
		ver := obj.GetLabels()[helm.ReleaseVersion]

		if gen == "" || ver == "" {
			return false, nil
		}

		if ver != c.Version() {
			return true, nil
		}

		g, err := strconv.Atoi(gen)
		if err != nil {
			return false, fmt.Errorf("cannot determine release generation: %w", err)
		}

		return rc.Resource.Generation > int64(g), nil
	})
	if err != nil {
		return fmt.Errorf("cannot run gc: %w", err)
	}

	a.l.Info("gc", "deleted", deleted)

	return nil
}

func (a *GCAction) Cleanup(_ context.Context, _ *ReconciliationRequest) error {
	return nil
}
