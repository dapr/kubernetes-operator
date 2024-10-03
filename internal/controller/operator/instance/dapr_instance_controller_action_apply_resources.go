package instance

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	ctrlCli "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dapr/kubernetes-operator/pkg/gvks"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/dapr/kubernetes-operator/pkg/controller/predicates"

	"github.com/dapr/kubernetes-operator/pkg/controller"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	daprApi "github.com/dapr/kubernetes-operator/api/operator/v1beta1"
	"github.com/dapr/kubernetes-operator/pkg/controller/client"
	"github.com/dapr/kubernetes-operator/pkg/helm"
	"github.com/dapr/kubernetes-operator/pkg/resources"
)

func NewApplyResourcesAction(l logr.Logger) Action {
	action := ApplyResourcesAction{
		l:             l.WithName("action").WithName("apply").WithName("resources"),
		subscriptions: make(map[string]struct{}),
	}

	return &action
}

type ApplyResourcesAction struct {
	l             logr.Logger
	subscriptions map[string]struct{}
}

func (a *ApplyResourcesAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	return b, nil
}

func (a *ApplyResourcesAction) Run(ctx context.Context, rc *ReconciliationRequest) error {
	c, err := rc.Chart(ctx)
	if err != nil {
		return fmt.Errorf("cannot load chart: %w", err)
	}

	items, err := c.Render(
		ctx,
		rc.Resource.Name,
		rc.Resource.Spec.Deployment.Namespace,
		int(rc.Resource.Generation),
		rc.Helm.ChartValues,
	)

	if err != nil {
		return fmt.Errorf("cannot render a chart: %w", err)
	}

	sort.Slice(items, func(i int, j int) bool {
		return resources.Ref(&items[i]) < resources.Ref(&items[j])
	})

	force := rc.Resource.Generation != rc.Resource.Status.ObservedGeneration || !helm.IsSameChart(c, rc.Resource.Status.Chart)

	if force {
		rc.Reconciler.Event(
			rc.Resource,
			corev1.EventTypeNormal,
			"RenderFullHelmTemplate",
			fmt.Sprintf("Render full Helm template (observedGeneration: %d, generation: %d, installedChart: %v, chart: %v)",
				rc.Resource.Status.ObservedGeneration,
				rc.Resource.Generation,
				rc.Resource.Status.Chart,
				c.Spec()),
		)
	}

	for _, obj := range items {
		resources.Labels(&obj, map[string]string{
			helm.ReleaseGeneration: strconv.FormatInt(rc.Resource.Generation, 10),
			helm.ReleaseName:       rc.Resource.Name,
			helm.ReleaseVersion:    c.Version(),
		})

		if err = a.apply(ctx, rc, &obj, force || !a.installOnly(&obj)); err != nil {
			return err
		}
	}

	return nil
}

func (a *ApplyResourcesAction) Cleanup(_ context.Context, _ *ReconciliationRequest) error {
	return nil
}

func (a *ApplyResourcesAction) apply(ctx context.Context, rc *ReconciliationRequest, in *unstructured.Unstructured, force bool) error {
	obj := in.DeepCopy()
	obj.SetNamespace(rc.Resource.Spec.Deployment.Namespace)
	obj.SetOwnerReferences(resources.OwnerReferences(rc.Resource))

	dc, err := rc.Client.Dynamic(obj)
	if err != nil {
		return fmt.Errorf("cannot create dynamic client: %w", err)
	}

	if dc.Scope() == client.ResourceScopeCluster {
		obj.SetNamespace("")
	}

	if err := a.watchResource(rc, obj); err != nil {
		return err
	}

	if !force {
		exists, err := a.exists(ctx, rc, in)

		switch {
		case err != nil:
			return err
		case !exists:
			break
		default:
			//
			// Every time the template is rendered, the helm function genSignedCert kicks in and
			// re-generated certs which causes a number os side effects and makes the set-up quite
			// unstable. As consequence some resources are not meant to be watched and re-created
			// unless the Dapr CR generation changes (which means the Spec has changed) or the
			// resource impacted by the genSignedCert hook is deleted.
			//
			// Ideally on OpenShift it would be good to leverage the service serving certificates
			// capability.
			//
			// Related info:
			// - https://docs.openshift.com/container-platform/4.13/security/certificates/service-serving-certificate.html
			// - https://github.com/dapr/dapr/issues/3968
			// - https://github.com/dapr/dapr/issues/6500
			//
			a.l.Info("run",
				"apply", "false",
				"ref", resources.Ref(obj),
				"gen", in.GetLabels()[helm.ReleaseGeneration],
				"version", in.GetLabels()[helm.ReleaseVersion],
				"scope", dc.Scope(),
				"reason", "resource marked as install-only")

			return nil
		}
	}

	_, err = dc.Apply(ctx, obj.GetName(), obj, metav1.ApplyOptions{
		FieldManager: controller.FieldManager,
		Force:        true,
	})

	if err != nil {
		return fmt.Errorf("cannot patch object %s: %w", resources.Ref(obj), err)
	}

	a.l.Info("run",
		"apply", "true",
		"ref", resources.Ref(obj),
		"gen", in.GetLabels()[helm.ReleaseGeneration],
		"version", in.GetLabels()[helm.ReleaseVersion],
		"scope", dc.Scope())

	return nil
}

func (a *ApplyResourcesAction) watchResource(rc *ReconciliationRequest, obj *unstructured.Unstructured) error {
	r := resources.GvkRef(obj)

	if _, ok := a.subscriptions[r]; ok {
		return nil
	}

	if _, ok := a.subscriptions[r]; !ok {
		var err error

		if a.watchStatus(obj) {
			a.l.Info("watch", "ref", r, "meta-only", false)

			err = rc.Reconciler.Watch(
				obj,
				rc.Reconciler.EnqueueRequestForOwner(&daprApi.DaprInstance{}, handler.OnlyControllerOwner()),
				dependantWithLabels(
					predicates.WithWatchUpdate(!a.installOnly(obj)),
					predicates.WithWatchDeleted(true),
					predicates.WithWatchStatus(true),
				),
			)
		} else {
			a.l.Info("watch", "ref", r, "meta-only", true)

			po := metav1.PartialObjectMetadata{}
			po.SetGroupVersionKind(obj.GroupVersionKind())

			err = rc.Reconciler.Watch(
				&po,
				rc.Reconciler.EnqueueRequestForOwner(&daprApi.DaprInstance{}, handler.OnlyControllerOwner()),
				partialDependantWithLabels(
					predicates.PartialWatchUpdate(!a.installOnly(obj)),
					predicates.PartialWatchDeleted(true),
				),
			)
		}

		if err != nil {
			return err
		}

		a.subscriptions[r] = struct{}{}
	}

	return nil
}

func (a *ApplyResourcesAction) watchStatus(obj ctrlCli.Object) bool {
	in := obj.GetObjectKind().GroupVersionKind()

	switch {
	case in == gvks.Deployment:
		return true
	case in == gvks.StatefulSet:
		return true
	default:
		return false
	}
}

func (a *ApplyResourcesAction) installOnly(obj ctrlCli.Object) bool {
	in := obj.GetObjectKind().GroupVersionKind()

	switch {
	case in == gvks.Secret:
		return true
	case in == gvks.MutatingWebhookConfiguration:
		return true
	case in == gvks.CustomResourceDefinition:
		return true
	default:
		return false
	}
}

func (a *ApplyResourcesAction) exists(ctx context.Context, rc *ReconciliationRequest, in ctrlCli.Object) (bool, error) {
	var obj ctrlCli.Object

	if !a.watchStatus(in) {
		p := metav1.PartialObjectMetadata{}
		p.SetGroupVersionKind(in.GetObjectKind().GroupVersionKind())

		obj = &p
	} else {
		p := unstructured.Unstructured{}
		p.SetGroupVersionKind(in.GetObjectKind().GroupVersionKind())

		obj = &p
	}

	err := rc.Client.Get(ctx, ctrlCli.ObjectKeyFromObject(in), obj)
	if k8serrors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("cannot get object %s: %w", resources.Ref(in), err)
	}

	return true, nil
}
