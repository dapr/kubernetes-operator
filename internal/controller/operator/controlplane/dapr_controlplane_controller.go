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

package controlplane

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/dapr/kubernetes-operator/pkg/openshift"

	"github.com/dapr/kubernetes-operator/pkg/helm"

	"k8s.io/client-go/tools/record"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	daprApi "github.com/dapr/kubernetes-operator/api/operator/v1alpha1"
	"github.com/dapr/kubernetes-operator/pkg/controller"
	"github.com/dapr/kubernetes-operator/pkg/controller/client"
	"github.com/go-logr/logr"

	ctrlRt "sigs.k8s.io/controller-runtime"
	ctrl "sigs.k8s.io/controller-runtime/pkg/controller"
)

func NewReconciler(ctx context.Context, manager ctrlRt.Manager, o helm.Options) (*Reconciler, error) {
	c, err := client.NewClient(manager.GetConfig(), manager.GetScheme(), manager.GetClient())
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	rec := Reconciler{}
	rec.l = ctrlRt.Log.WithName("dapr-controlplane-controller")
	rec.client = c
	rec.Scheme = manager.GetScheme()
	rec.ClusterType = controller.ClusterTypeVanilla
	rec.manager = manager
	rec.recorder = manager.GetEventRecorderFor(controller.FieldManager)

	isOpenshift, err := openshift.IsOpenShift(c.Discovery)
	if err != nil {
		//nolint: wrapcheck
		return nil, err
	}

	if isOpenshift {
		rec.ClusterType = controller.ClusterTypeOpenShift
	}

	rec.actions = append(rec.actions, NewApplyAction(rec.l))
	rec.actions = append(rec.actions, NewStatusAction(rec.l))

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

// +kubebuilder:rbac:groups=operator.dapr.io,resources=daprcontrolplanes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.dapr.io,resources=daprcontrolplanes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.dapr.io,resources=daprcontrolplanes/finalizers,verbs=update
// +kubebuilder:rbac:groups=operator.dapr.io,resources=daprinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.dapr.io,resources=daprinstances/status,verbs=get

type Reconciler struct {
	client      *client.Client
	Scheme      *runtime.Scheme
	ClusterType controller.ClusterType
	actions     []Action
	l           logr.Logger
	c           *chart.Chart
	manager     ctrlRt.Manager
	controller  ctrl.Controller
	recorder    record.EventRecorder
}

func (r *Reconciler) Client() *client.Client {
	return r.client
}

func (r *Reconciler) init(ctx context.Context) error {
	c := ctrlRt.NewControllerManagedBy(r.manager)

	c = c.For(&daprApi.DaprControlPlane{}, builder.WithPredicates(
		predicate.Or(
			predicate.GenerationChangedPredicate{},
		)))

	for i := range r.actions {
		b, err := r.actions[i].Configure(ctx, r.Client(), c)
		if err != nil {
			//nolint:wrapcheck
			return err
		}

		c = b
	}

	objRec := reconcile.AsReconciler[*daprApi.DaprControlPlane](r.manager.GetClient(), r)

	ct, err := c.Build(objRec)
	if err != nil {
		return fmt.Errorf("failure building the application controller for DaprControlPlane resource: %w", err)
	}

	r.controller = ct

	return nil
}

func (r *Reconciler) Event(object runtime.Object, eventType string, reason string, message string) {
	r.recorder.Event(
		object,
		eventType,
		reason,
		message,
	)
}
