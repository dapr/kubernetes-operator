/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package instance

import (
	"context"
	"fmt"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/openshift"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/helm"

	"k8s.io/client-go/tools/record"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	daprvApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller"
	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/client"
	"github.com/go-logr/logr"

	ctrlRt "sigs.k8s.io/controller-runtime"
	ctrlCli "sigs.k8s.io/controller-runtime/pkg/client"
	ctrl "sigs.k8s.io/controller-runtime/pkg/controller"
)

func NewReconciler(ctx context.Context, manager ctrlRt.Manager, o helm.Options) (*Reconciler, error) {
	c, err := client.NewClient(manager.GetConfig(), manager.GetScheme(), manager.GetClient())
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	rec := Reconciler{}
	rec.l = ctrlRt.Log.WithName("dapr-instance-controller")
	rec.Client = c
	rec.Scheme = manager.GetScheme()
	rec.ClusterType = controller.ClusterTypeVanilla
	rec.manager = manager
	rec.recorder = manager.GetEventRecorderFor(controller.FieldManager)

	isOpenshift, err := openshift.IsOpenShift(c.Discovery)
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	if isOpenshift {
		rec.ClusterType = controller.ClusterTypeOpenShift
	}

	rec.actions = append(rec.actions, NewChartAction(rec.l))
	rec.actions = append(rec.actions, NewApplyCRDsAction(rec.l))
	rec.actions = append(rec.actions, NewApplyResourcesAction(rec.l))
	rec.actions = append(rec.actions, NewConditionsAction(rec.l))

	hc, err := loader.Load(o.ChartsDir)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart from dir %s: %w", o.ChartsDir, err)
	}

	rec.c = hc
	if rec.c.Values == nil {
		rec.c.Values = make(map[string]interface{})
	}

	err = rec.init(ctx)
	if err != nil {
		return nil, err
	}

	return &rec, nil
}

// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;create;update;patch
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=*
// +kubebuilder:rbac:groups=operator.dapr.io,resources=daprinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.dapr.io,resources=daprinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.dapr.io,resources=daprinstances/finalizers,verbs=update
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=*
// +kubebuilder:rbac:groups="",resources=events,verbs=*
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=*
// +kubebuilder:rbac:groups="",resources=secrets,verbs=*
// +kubebuilder:rbac:groups="",resources=services,verbs=*
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=*
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=*
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=*
// +kubebuilder:rbac:groups=dapr.io,resources=components,verbs=*
// +kubebuilder:rbac:groups=dapr.io,resources=components/status,verbs=*
// +kubebuilder:rbac:groups=dapr.io,resources=components/finalizers,verbs=*
// +kubebuilder:rbac:groups=dapr.io,resources=configurations,verbs=*
// +kubebuilder:rbac:groups=dapr.io,resources=configurations/status,verbs=*
// +kubebuilder:rbac:groups=dapr.io,resources=configurations/finalizers,verbs=*
// +kubebuilder:rbac:groups=dapr.io,resources=resiliencies,verbs=*
// +kubebuilder:rbac:groups=dapr.io,resources=resiliencies/status,verbs=*
// +kubebuilder:rbac:groups=dapr.io,resources=resiliencies/finalizers,verbs=*
// +kubebuilder:rbac:groups=dapr.io,resources=subscriptions,verbs=*
// +kubebuilder:rbac:groups=dapr.io,resources=subscriptions/status,verbs=*
// +kubebuilder:rbac:groups=dapr.io,resources=subscriptions/finalizers,verbs=*

type Reconciler struct {
	*client.Client

	Scheme      *runtime.Scheme
	ClusterType controller.ClusterType
	actions     []Action
	l           logr.Logger
	c           *chart.Chart
	manager     ctrlRt.Manager
	controller  ctrl.Controller
	recorder    record.EventRecorder
}

func (r *Reconciler) init(ctx context.Context) error {
	c := ctrlRt.NewControllerManagedBy(r.manager)

	c = c.For(&daprvApi.DaprInstance{}, builder.WithPredicates(
		predicate.Or(
			predicate.GenerationChangedPredicate{},
		)))

	for i := range r.actions {
		b, err := r.actions[i].Configure(ctx, r.Client, c)
		if err != nil {
			//nolint:wrapcheck
			return err
		}

		c = b
	}

	ct, err := c.Build(r)
	if err != nil {
		return fmt.Errorf("failure building the application controller for DaprInstance resource: %w", err)
	}

	r.controller = ct

	return nil
}

func (r *Reconciler) Watch(obj ctrlCli.Object, eh handler.EventHandler, predicates ...predicate.Predicate) error {
	err := r.controller.Watch(
		source.Kind(
			r.manager.GetCache(),
			obj,
			eh,
			predicates...,
		),
	)

	if err != nil {
		return fmt.Errorf(
			"error configuring watcher for resource %s:%s, reson: %w",
			obj.GetObjectKind().GroupVersionKind().String(),
			obj.GetName(),
			err)
	}

	return nil
}

func (r *Reconciler) EnqueueRequestForOwner(owner ctrlCli.Object, opts ...handler.OwnerOption) handler.EventHandler {
	return handler.EnqueueRequestForOwner(
		r.manager.GetScheme(),
		r.manager.GetRESTMapper(),
		owner,
		opts...,
	)
}

func (r *Reconciler) EnqueueRequestsFromMapFunc(fn func(context.Context, ctrlCli.Object) []reconcile.Request) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(fn)
}

func (r *Reconciler) Event(object runtime.Object, eventType string, reason string, message string) {
	r.recorder.Event(
		object,
		eventType,
		reason,
		message,
	)
}
