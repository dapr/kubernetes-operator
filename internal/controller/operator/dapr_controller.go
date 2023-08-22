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

package operator

import (
	"context"

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

func NewReconciler(ctx context.Context, manager ctrlRt.Manager, o HelmOptions) (*Reconciler, error) {
	c, err := client.NewClient(manager.GetConfig(), manager.GetScheme(), manager.GetClient())
	if err != nil {
		return nil, err
	}

	rec := Reconciler{}
	rec.l = ctrlRt.Log.WithName("controller")
	rec.Client = c
	rec.Scheme = manager.GetScheme()
	rec.ClusterType = controller.ClusterTypeVanilla
	rec.manager = manager
	rec.recorder = manager.GetEventRecorderFor(DaprFieldManager)

	isOpenshift, err := c.IsOpenShift()
	if err != nil {
		return nil, err
	}
	if isOpenshift {
		rec.ClusterType = controller.ClusterTypeOpenShift
	}

	rec.actions = append(rec.actions, NewApplyAction())

	hc, err := loader.Load(o.ChartsDir)
	if err != nil {
		return nil, err
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

//+kubebuilder:rbac:groups=operator.dapr.io,resources=daprcontrolplanes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.dapr.io,resources=daprcontrolplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.dapr.io,resources=daprcontrolplanes/finalizers,verbs=update
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=*
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=*
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=*
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=*
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=*
//+kubebuilder:rbac:groups="",resources=events,verbs=*
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=*
//+kubebuilder:rbac:groups="",resources=secrets,verbs=*
//+kubebuilder:rbac:groups="",resources=services,verbs=*
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=*
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=*
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=components,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=components/status,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=components/finalizers,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=configurations,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=configurations/status,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=configurations/finalizers,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=resiliencies,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=resiliencies/status,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=resiliencies/finalizers,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=subscriptions,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=subscriptions/status,verbs=*
//+kubebuilder:rbac:groups=dapr.io,resources=subscriptions/finalizers,verbs=*

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

	// TODO: as today, the controller can handle multiple Dapr CR however, the Dapr operator does
	//       not seem to be designed to handle multiple installations on the same cluster hence
	//       we must discuss if we want to limit to a single CR or even remove the Dapr CR and
	//       use a simple ConfigMap (which should be less practical as having a place like the
	//       status field where to report what's going on is highly desirable.
	c = c.For(&daprvApi.DaprControlPlane{}, builder.WithPredicates(
		predicate.Or(
			predicate.GenerationChangedPredicate{},
		)))

	for i := range r.actions {
		b, err := r.actions[i].Configure(ctx, r.Client, c)
		if err != nil {
			return err
		}

		c = b
	}

	ct, err := c.Build(r)
	if err != nil {
		return err
	}

	r.controller = ct

	return nil
}

func (r *Reconciler) Watch(obj ctrlCli.Object, eh handler.EventHandler, predicates ...predicate.Predicate) error {
	return r.controller.Watch(
		source.Kind(r.manager.GetCache(), obj),
		eh,
		predicates...)
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
