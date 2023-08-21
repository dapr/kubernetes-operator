package operator

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/gc"
	corev1 "k8s.io/api/core/v1"

	daprApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/pointer"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/resources"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/client"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/helm"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

func NewApplyAction() Action {
	return &ApplyAction{
		engine:        helm.NewEngine(),
		l:             ctrl.Log.WithName("action").WithName("apply"),
		subscriptions: make(map[string]struct{}),
		gc:            gc.New(),
	}
}

type ApplyAction struct {
	engine        *helm.Engine
	gc            *gc.GC
	l             logr.Logger
	subscriptions map[string]struct{}
}

func (a *ApplyAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	return b, nil
}

func (a *ApplyAction) Run(ctx context.Context, rc *ReconciliationRequest) error {
	items, err := a.engine.Render(rc.Chart, rc.Resource, rc.Overrides)
	if err != nil {
		return errors.Wrap(err, "cannot render a chart")
	}

	// TODO: this must be ordered by priority/relations
	sort.Slice(items, func(i int, j int) bool {
		istr := items[i].GroupVersionKind().Kind + ":" + items[i].GetName()
		jstr := items[j].GroupVersionKind().Kind + ":" + items[j].GetName()

		return istr < jstr
	})

	reinstall := rc.Resource.Generation != rc.Resource.Status.ObservedGeneration

	if reinstall {
		rc.Reconciler.Event(
			rc.Resource,
			corev1.EventTypeNormal,
			"RenderFullHelmTemplate",
			fmt.Sprintf("Render full Helm template as Dapr spec changed (observedGeneration: %d, generation: %d)",
				rc.Resource.Status.ObservedGeneration,
				rc.Resource.Generation),
		)
	}

	for i := range items {
		obj := items[i]
		gvk := obj.GroupVersionKind()
		installOnly := a.installOnly(gvk)

		if reinstall {
			installOnly = false
		}

		dc, err := rc.Client.Dynamic(rc.Resource.Namespace, &obj)
		if err != nil {
			return errors.Wrap(err, "cannot create dynamic client")
		}

		resources.Labels(&obj, map[string]string{
			DaprReleaseGeneration: strconv.FormatInt(rc.Resource.Generation, 10),
			DaprReleaseName:       rc.Resource.Name,
			DaprReleaseNamespace:  rc.Resource.Namespace,
		})

		switch dc.(type) {

		//
		// NamespacedResource: in this case, filtering with ownership can be implemented
		// as all the namespaced resources created by this controller have the Dapr CR as
		// an owner
		//
		case *client.NamespacedResource:
			obj.SetOwnerReferences(resources.OwnerReferences(rc.Resource))
			obj.SetNamespace(rc.Resource.Namespace)

			r := gvk.GroupVersion().String() + ":" + gvk.Kind

			if _, ok := a.subscriptions[r]; !ok {
				err = rc.Reconciler.Watch(
					&obj,
					rc.Reconciler.EnqueueRequestForOwner(&daprApi.DaprControlPlane{}, handler.OnlyControllerOwner()),
					dependantWithLabels(
						a.watchForUpdates(gvk),
						true),
				)

				if err != nil {
					return err
				}

				a.subscriptions[r] = struct{}{}
			}

		//
		// ClusteredResource: in this case, ownership based filtering is not supported
		// as you cannot have a non namespaced owner. For such reason, the resource for
		// which a reconcile should be triggered can be identified by using the labels
		// added by the controller to all the generated resources
		//
		//    daprs.dapr.io/resource.namespace = ${namespace}
		//    daprs.dapr.io/resource.name =${name}
		//
		case *client.ClusteredResource:
			r := gvk.GroupVersion().String() + ":" + gvk.Kind

			if _, ok := a.subscriptions[r]; !ok {
				err = rc.Reconciler.Watch(
					&obj,
					rc.Reconciler.EnqueueRequestsFromMapFunc(labelsToRequest),
					dependantWithLabels(
						a.watchForUpdates(gvk),
						true),
				)

				if err != nil {
					return err
				}

				a.subscriptions[r] = struct{}{}
			}
		}

		if installOnly {
			old, err := dc.Get(ctx, obj.GetName(), metav1.GetOptions{})
			if err != nil {
				if !k8serrors.IsNotFound(err) {
					return errors.Wrapf(err, "cannot get object %s", resources.Ref(&obj))
				}
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
					"ref", resources.Ref(&obj),
					"reason", "resource marked as install-only")

				continue
			}
		}

		_, err = dc.Apply(ctx, obj.GetName(), &obj, metav1.ApplyOptions{
			FieldManager: DaprFieldManager,
			Force:        true,
		})

		if err != nil {
			return errors.Wrapf(err, "cannot patch object %s", resources.Ref(&obj))
		}

		a.l.Info("run",
			"apply", "true",
			"ref", resources.Ref(&obj))
	}

	//
	// in case of a re-installation all the resources get re-rendered which means some of them
	// may become obsolete (i.e. if some resources are moved from cluster to namespace scope)
	// hence a sort of "garbage collector task" must be executed.
	//
	// The logic of the task it to delete all the resources that have a generation older than
	// current CR one, which is propagated by the controller to all the rendered resources in
	// the for of a labels:
	//
	// - daprs.tools.dapr.io/release.generation
	//
	if reinstall {
		s, err := gcSelector(rc)
		if err != nil {
			return errors.Wrap(err, "cannot compute gc selector")
		}

		deleted, err := a.gc.Run(ctx, rc.Resource.Namespace, rc.Client, s)
		if err != nil {
			return errors.Wrap(err, "cannot run gc")
		}

		a.l.Info("gc", "deleted", deleted)
	}

	return nil
}

func (a *ApplyAction) Cleanup(ctx context.Context, rc *ReconciliationRequest) error {
	items, err := a.engine.Render(rc.Chart, rc.Resource, rc.Overrides)
	if err != nil {
		return errors.Wrap(err, "cannot render a chart")
	}

	for i := range items {
		obj := items[i]

		dc, err := rc.Client.Dynamic(rc.Resource.Namespace, &obj)
		if err != nil {
			return errors.Wrap(err, "cannot create dynamic client")
		}

		// Delete clustered resources
		if _, ok := dc.(*client.ClusteredResource); ok {
			err := dc.Delete(ctx, obj.GetName(), metav1.DeleteOptions{
				PropagationPolicy: pointer.Any(metav1.DeletePropagationForeground),
			})

			if err != nil && !k8serrors.IsNotFound(err) {
				return errors.Wrapf(err, "cannot delete object %s", resources.Ref(&obj))
			}

			a.l.Info("delete", "ref", resources.Ref(&obj))
		}
	}

	return nil
}

func (a *ApplyAction) watchForUpdates(gvk schema.GroupVersionKind) bool {
	if gvk.Group == "" && gvk.Version == "v1" && gvk.Kind == "Secret" {
		return false
	}
	if gvk.Group == "admissionregistration.k8s.io" && gvk.Version == "v1" && gvk.Kind == "MutatingWebhookConfiguration" {
		return false
	}

	return true
}

func (a *ApplyAction) installOnly(gvk schema.GroupVersionKind) bool {
	if gvk.Group == "" && gvk.Version == "v1" && gvk.Kind == "Secret" {
		return true
	}
	if gvk.Group == "admissionregistration.k8s.io" && gvk.Version == "v1" && gvk.Kind == "MutatingWebhookConfiguration" {
		return true
	}

	return false
}
