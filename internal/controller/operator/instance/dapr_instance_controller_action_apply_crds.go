package instance

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/helm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/client"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/resources"
	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

func NewApplyCRDsAction(l logr.Logger) Action {
	action := ApplyCRDsAction{
		l:       l.WithName("action").WithName("apply").WithName("crds"),
		decoder: k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme),
	}

	return &action
}

type ApplyCRDsAction struct {
	l       logr.Logger
	decoder runtime.Serializer
}

func (a *ApplyCRDsAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	return b, nil
}

func (a *ApplyCRDsAction) Run(ctx context.Context, rc *ReconciliationRequest) error {
	if rc.Resource.Generation == rc.Resource.Status.ObservedGeneration {
		return nil
	}

	crds := rc.Chart.CRDObjects()

	sort.Slice(crds, func(i, j int) bool {
		return crds[i].Name < crds[j].Name
	})

	for _, crd := range crds {
		items, err := resources.Decode(a.decoder, crd.File.Data)
		if err != nil {
			return fmt.Errorf("cannot decode CRD %s: %w", crd.Name, err)
		}

		for i := range items {
			obj := items[i]

			dc, err := rc.Client.Dynamic(rc.Resource.Namespace, &obj)
			if err != nil {
				return fmt.Errorf("cannot create dynamic client: %w", err)
			}

			resources.Labels(&obj, map[string]string{
				helm.ReleaseGeneration: strconv.FormatInt(rc.Resource.Generation, 10),
				helm.ReleaseName:       rc.Resource.Name,
				helm.ReleaseNamespace:  rc.Resource.Namespace,
			})

			apply := rc.Resource.Generation != rc.Resource.Status.ObservedGeneration
			_, err = dc.Get(ctx, obj.GetName(), metav1.GetOptions{})

			if err != nil && !k8serrors.IsNotFound(err) {
				return fmt.Errorf("cannot determine if CRD %s exists: %w", resources.Ref(&obj), err)
			}
			if err != nil && k8serrors.IsNotFound(err) {
				apply = true
			}

			if !apply {
				a.l.Info("run",
					"apply", "false",
					"gen", rc.Resource.Generation,
					"ref", resources.Ref(&obj),
					"generation-changed", rc.Resource.Generation != rc.Resource.Status.ObservedGeneration,
					"not-found", k8serrors.IsNotFound(err))

				continue
			}

			_, err = dc.Apply(ctx, obj.GetName(), &obj, metav1.ApplyOptions{
				FieldManager: controller.FieldManager,
				Force:        true,
			})

			if err != nil {
				return fmt.Errorf("cannot apply CRD %s: %w", resources.Ref(&obj), err)
			}

			a.l.Info("run",
				"apply", "true",
				"gen", rc.Resource.Generation,
				"ref", resources.Ref(&obj))

		}
	}

	// invalidate the client so it gets aware of the new CRDs
	rc.Client.Invalidate()

	return nil
}

func (a *ApplyCRDsAction) Cleanup(_ context.Context, _ *ReconciliationRequest) error {
	return nil
}
