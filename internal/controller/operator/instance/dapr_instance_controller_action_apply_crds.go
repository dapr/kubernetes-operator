package instance

import (
	"context"
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/dapr/kubernetes-operator/pkg/controller"
	"github.com/dapr/kubernetes-operator/pkg/helm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/dapr/kubernetes-operator/pkg/controller/client"
	"github.com/dapr/kubernetes-operator/pkg/resources"
	"github.com/go-logr/logr"

	"sigs.k8s.io/controller-runtime/pkg/builder"
)

func NewApplyCRDsAction(l logr.Logger) Action {
	action := ApplyCRDsAction{
		l: l.WithName("action").WithName("apply").WithName("crds"),
	}

	return &action
}

type ApplyCRDsAction struct {
	l logr.Logger
}

func (a *ApplyCRDsAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	return b, nil
}

func (a *ApplyCRDsAction) Run(ctx context.Context, rc *ReconciliationRequest) error {
	if rc.Resource.Generation == rc.Resource.Status.ObservedGeneration {
		return nil
	}

	c, err := rc.Chart(ctx)
	if err != nil {
		return fmt.Errorf("cannot load chart: %w", err)
	}

	crds, err := c.CRDObjects()
	if err != nil {
		return fmt.Errorf("cannot load CRDs: %w", err)
	}

	invalidate := false
	force := rc.Resource.Generation != rc.Resource.Status.ObservedGeneration || !helm.IsSameChart(c, rc.Resource.Status.Chart)

	for _, crd := range crds {
		resources.Labels(&crd, map[string]string{
			helm.ReleaseGeneration: strconv.FormatInt(rc.Resource.Generation, 10),
			helm.ReleaseName:       rc.Resource.Name,
			helm.ReleaseNamespace:  rc.Resource.Namespace,
			helm.ReleaseVersion:    c.Version(),
		})

		applied, err := a.apply(ctx, rc, &crd, force)
		if err != nil {
			return err
		}

		if applied {
			invalidate = true
		}
	}

	if invalidate {
		// invalidate the client so it gets aware of the new CRDs
		rc.Client.Invalidate()
	}

	return nil
}

func (a *ApplyCRDsAction) Cleanup(_ context.Context, _ *ReconciliationRequest) error {
	return nil
}

func (a *ApplyCRDsAction) apply(ctx context.Context, rc *ReconciliationRequest, crd *unstructured.Unstructured, apply bool) (bool, error) {
	dc, err := rc.Client.Dynamic(rc.Resource.Namespace, crd)
	if err != nil {
		return false, fmt.Errorf("cannot create dynamic client: %w", err)
	}

	_, err = dc.Get(ctx, crd.GetName(), metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return false, fmt.Errorf("cannot determine if CRD %s exists: %w", resources.Ref(crd), err)
	}

	if k8serrors.IsNotFound(err) {
		apply = true
	}

	if !apply {
		a.l.Info("run",
			"apply", "false",
			"gen", rc.Resource.Generation,
			"ref", resources.Ref(crd))

		return false, nil
	}

	_, err = dc.Apply(ctx, crd.GetName(), crd, metav1.ApplyOptions{
		FieldManager: controller.FieldManager,
		Force:        true,
	})

	if err != nil {
		return false, fmt.Errorf("cannot apply CRD %s: %w", resources.Ref(crd), err)
	}

	a.l.Info("run",
		"apply", "true",
		"gen", rc.Resource.Generation,
		"ref", resources.Ref(crd))

	return true, nil
}
