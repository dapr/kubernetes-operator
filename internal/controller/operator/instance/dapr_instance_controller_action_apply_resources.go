package instance

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/dapr/kubernetes-operator/pkg/controller/predicates"

	"github.com/dapr/kubernetes-operator/pkg/controller"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	daprApi "github.com/dapr/kubernetes-operator/api/operator/v1alpha1"
	"github.com/dapr/kubernetes-operator/pkg/controller/client"
	"github.com/dapr/kubernetes-operator/pkg/helm"
	"github.com/dapr/kubernetes-operator/pkg/pointer"
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

	items, err := c.Render(ctx, rc.Resource.Name, rc.Resource.Namespace, int(rc.Resource.Generation), rc.Helm.ChartValues)
	if err != nil {
		return fmt.Errorf("cannot render a chart: %w", err)
	}

	// TODO: this must be ordered by priority/relations
	sort.Slice(items, func(i int, j int) bool {
		istr := items[i].GroupVersionKind().Kind + ":" + items[i].GetName()
		jstr := items[j].GroupVersionKind().Kind + ":" + items[j].GetName()

		return istr < jstr
	})

	installedVersion := ""
	if rc.Resource.Status.Chart != nil {
		installedVersion = rc.Resource.Status.Chart.Version
	}

	force := rc.Resource.Generation != rc.Resource.Status.ObservedGeneration || c.Version() != installedVersion

	if force {
		rc.Reconciler.Event(
			rc.Resource,
			corev1.EventTypeNormal,
			"RenderFullHelmTemplate",
			fmt.Sprintf("Render full Helm template (observedGeneration: %d, generation: %d, installedChartVersion: %s, chartVersion: %s)",
				rc.Resource.Status.ObservedGeneration,
				rc.Resource.Generation,
				installedVersion,
				c.Version()),
		)
	}

	for _, obj := range items {
		resources.Labels(&obj, map[string]string{
			helm.ReleaseGeneration: strconv.FormatInt(rc.Resource.Generation, 10),
			helm.ReleaseName:       rc.Resource.Name,
			helm.ReleaseNamespace:  rc.Resource.Namespace,
			helm.ReleaseVersion:    c.Version(),
		})

		gvk := obj.GroupVersionKind()

		if !force {
			force = !a.installOnly(gvk)
		}

		err = a.apply(ctx, rc, &obj, force)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *ApplyResourcesAction) Cleanup(ctx context.Context, rc *ReconciliationRequest) error {
	c, err := rc.Chart(ctx)
	if err != nil {
		return fmt.Errorf("cannot load chart: %w", err)
	}

	items, err := c.Render(ctx, rc.Resource.Name, rc.Resource.Namespace, int(rc.Resource.Generation), rc.Helm.ChartValues)
	if err != nil {
		return fmt.Errorf("cannot render a chart: %w", err)
	}

	for i := range items {
		obj := items[i]

		dc, err := rc.Client.Dynamic(rc.Resource.Namespace, &obj)
		if err != nil {
			return fmt.Errorf("cannot create dynamic client: %w", err)
		}

		// Delete clustered resources
		if _, ok := dc.(*client.ClusteredResource); ok {
			err := dc.Delete(ctx, obj.GetName(), metav1.DeleteOptions{
				PropagationPolicy: pointer.Any(metav1.DeletePropagationForeground),
			})

			if err != nil && !k8serrors.IsNotFound(err) {
				return fmt.Errorf("cannot delete object %s: %w", resources.Ref(&obj), err)
			}

			a.l.Info("delete", "ref", resources.Ref(&obj))
		}
	}

	return nil
}

func (a *ApplyResourcesAction) watchForUpdates(gvk schema.GroupVersionKind) bool {
	if gvk.Group == "" && gvk.Version == "v1" && gvk.Kind == "Secret" {
		return false
	}

	if gvk.Group == "admissionregistration.k8s.io" && gvk.Version == "v1" && gvk.Kind == "MutatingWebhookConfiguration" {
		return false
	}

	if gvk.Group == "apiextensions.k8s.io" && gvk.Version == "v1" && gvk.Kind == "CustomResourceDefinition" {
		return false
	}

	return true
}

func (a *ApplyResourcesAction) watchStatus(gvk schema.GroupVersionKind) bool {
	if gvk.Group == "apps" && gvk.Version == "v1" && gvk.Kind == "Deployment" {
		return true
	}

	return false
}

func (a *ApplyResourcesAction) installOnly(gvk schema.GroupVersionKind) bool {
	if gvk.Group == "" && gvk.Version == "v1" && gvk.Kind == "Secret" {
		return true
	}

	if gvk.Group == "admissionregistration.k8s.io" && gvk.Version == "v1" && gvk.Kind == "MutatingWebhookConfiguration" {
		return true
	}

	if gvk.Group == "apiextensions.k8s.io" && gvk.Version == "v1" && gvk.Kind == "CustomResourceDefinition" {
		return true
	}

	return false
}

//nolint:cyclop
func (a *ApplyResourcesAction) apply(ctx context.Context, rc *ReconciliationRequest, obj *unstructured.Unstructured, force bool) error {
	dc, err := rc.Client.Dynamic(rc.Resource.Namespace, obj)
	if err != nil {
		return fmt.Errorf("cannot create dynamic client: %w", err)
	}

	switch dc.(type) {
	//
	// NamespacedResource: in this case, filtering with ownership can be implemented
	// as all the namespaced resources created by this controller have the Dapr CR as
	// an owner
	//
	case *client.NamespacedResource:
		if err := a.watchNamespaceScopeResource(rc, obj); err != nil {
			return err
		}

	//
	// ClusteredResource: in this case, ownership based filtering is not supported
	// as you cannot have a non namespaced owner. For such reason, the resource for
	// which a reconcile should be triggered can be identified by using the labels
	// added by the controller to all the generated resources
	//
	//    helm.operator.dapr.io/resource.namespace = ${namespace}
	//    helm.operator.dapr.io/resource.name = ${name}
	//
	case *client.ClusteredResource:
		if err := a.watchClusterScopeResource(rc, obj); err != nil {
			return err
		}
	}

	if !force {
		old, err := dc.Get(ctx, obj.GetName(), metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return fmt.Errorf("cannot get object %s: %w", resources.Ref(obj), err)
		}

		if old != nil {
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
				"gen", rc.Resource.Generation,
				"ref", resources.Ref(obj),
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
		"gen", rc.Resource.Generation,
		"ref", resources.Ref(obj))

	return nil
}

func (a *ApplyResourcesAction) watchNamespaceScopeResource(rc *ReconciliationRequest, obj *unstructured.Unstructured) error {
	gvk := obj.GroupVersionKind()

	obj.SetOwnerReferences(resources.OwnerReferences(rc.Resource))
	obj.SetNamespace(rc.Resource.Namespace)

	r := gvk.GroupVersion().String() + ":" + gvk.Kind

	if _, ok := a.subscriptions[r]; ok {
		return nil
	}

	if _, ok := a.subscriptions[r]; !ok {
		a.l.Info("watch", "scope", "namespace", "ref", r)

		err := rc.Reconciler.Watch(
			obj,
			rc.Reconciler.EnqueueRequestForOwner(&daprApi.DaprInstance{}, handler.OnlyControllerOwner()),
			dependantWithLabels(
				predicates.WithWatchUpdate(a.watchForUpdates(gvk)),
				predicates.WithWatchDeleted(true),
				predicates.WithWatchStatus(a.watchStatus(gvk)),
			),
		)

		if err != nil {
			return err
		}

		a.subscriptions[r] = struct{}{}
	}

	return nil
}

func (a *ApplyResourcesAction) watchClusterScopeResource(rc *ReconciliationRequest, obj *unstructured.Unstructured) error {
	gvk := obj.GroupVersionKind()

	r := gvk.GroupVersion().String() + ":" + gvk.Kind

	if _, ok := a.subscriptions[r]; ok {
		return nil
	}

	if _, ok := a.subscriptions[r]; !ok {
		a.l.Info("watch", "scope", "cluster", "ref", r)

		err := rc.Reconciler.Watch(
			obj,
			rc.Reconciler.EnqueueRequestsFromMapFunc(labelsToRequest),
			dependantWithLabels(
				predicates.WithWatchUpdate(a.watchForUpdates(gvk)),
				predicates.WithWatchDeleted(true),
				predicates.WithWatchStatus(a.watchStatus(gvk)),
			),
		)

		if err != nil {
			return err
		}

		a.subscriptions[r] = struct{}{}
	}

	return nil
}
