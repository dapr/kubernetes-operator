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

	"k8s.io/client-go/tools/record"

	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	daprApi "github.com/dapr/kubernetes-operator/api/operator/v1alpha1"

	"github.com/dapr/kubernetes-operator/pkg/controller"
	"github.com/dapr/kubernetes-operator/pkg/controller/client"
	"github.com/dapr/kubernetes-operator/pkg/controller/reconciler"
	"github.com/dapr/kubernetes-operator/pkg/helm"
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
	rec.manager = manager
	rec.recorder = manager.GetEventRecorderFor(controller.FieldManager)

	rec.actions = append(rec.actions, NewApplyAction(rec.l))
	rec.actions = append(rec.actions, NewStatusAction(rec.l))

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
	client     *client.Client
	Scheme     *runtime.Scheme
	actions    []Action
	l          logr.Logger
	manager    ctrlRt.Manager
	controller ctrl.Controller
	recorder   record.EventRecorder
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

	// by default, the controller expect the DaprControlPlane resource to be created
	// in the same namespace where it runs, if not fallback to the default namespace
	rec := reconciler.BaseReconciler[*daprApi.DaprControlPlane]{
		Delegate:        r,
		Client:          r.client,
		Log:             log.FromContext(ctx),
		Name:            DaprControlPlaneResourceName,
		Namespace:       controller.OperatorNamespace(),
		FinalizerName:   DaprControlPlaneFinalizerName,
		FinalizerAction: r.Cleanup,
	}

	ct, err := c.Build(&rec)
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
