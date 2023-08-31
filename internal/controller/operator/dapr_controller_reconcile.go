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
	"fmt"
	"os"
	"sort"

	daprvApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ctrl "sigs.k8s.io/controller-runtime"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Reconciling", "resource", req.NamespacedName.String())

	rr := ReconciliationRequest{
		Client: r.Client,
		NamespacedName: types.NamespacedName{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		ClusterType: r.ClusterType,
		Reconciler:  r,
		Resource:    &daprvApi.DaprControlPlane{},
		Chart:       r.c,
		Overrides: map[string]interface{}{
			"dapr_operator":  map[string]interface{}{"runAsNonRoot": "true"},
			"dapr_placement": map[string]interface{}{"runAsNonRoot": "true"},
			"dapr_sentry":    map[string]interface{}{"runAsNonRoot": "true"},
			"dapr_dashboard": map[string]interface{}{"runAsNonRoot": "true"},
		},
	}

	err := r.Get(ctx, req.NamespacedName, rr.Resource)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// no CR found anymore, maybe deleted
			return ctrl.Result{}, nil
		}
	}

	// by default, the controller expect the DaprControlPlane resource to be created
	// in the same namespace where it runs, if not fallback to the default namespace
	// dapr-system
	ns := os.Getenv(DaprControlPlaneNamespaceEnv)
	if ns == "" {
		ns = DaprControlPlaneNamespaceDefault
	}

	if req.Name != DaprControlPlaneName || req.Namespace != ns {
		rr.Resource.Status.Phase = DaprPhaseError

		meta.SetStatusCondition(&rr.Resource.Status.Conditions, metav1.Condition{
			Type:   DaprConditionReconciled,
			Status: metav1.ConditionFalse,
			Reason: DaprConditionReasonUnsupportedConfiguration,
			Message: fmt.Sprintf(
				"Unsupported resource, the operator handles a single resource named %s in namespace %s",
				DaprControlPlaneName,
				ns),
		})

		err = r.Status().Update(ctx, rr.Resource)

		if err != nil && k8serrors.IsConflict(err) {
			l.Info(err.Error())
			return ctrl.Result{Requeue: true}, nil
		}

		return ctrl.Result{}, err
	}

	if rr.Resource.ObjectMeta.DeletionTimestamp.IsZero() {

		//
		// Add finalizer
		//

		if ctrlutil.AddFinalizer(rr.Resource, DaprFinalizerName) {
			if err := r.Update(ctx, rr.Resource); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, errors.Wrapf(err, "failure adding finalizer to connector cluster %s", req.NamespacedName)
			}
		}
	} else {

		//
		// Cleanup leftovers if needed
		//

		for i := len(r.actions) - 1; i >= 0; i-- {
			if err := r.actions[i].Cleanup(ctx, &rr); err != nil {
				return ctrl.Result{}, err
			}
		}

		//
		// Handle finalizer
		//

		if ctrlutil.RemoveFinalizer(rr.Resource, DaprFinalizerName) {
			if err := r.Update(ctx, rr.Resource); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, errors.Wrapf(err, "failure removing finalizer from %s", req.NamespacedName)
			}
		}

		return ctrl.Result{}, nil
	}

	//
	// Reconcile
	//

	reconcileCondition := metav1.Condition{
		Type:               DaprConditionReconciled,
		Status:             metav1.ConditionTrue,
		Reason:             "Reconciled",
		Message:            "Reconciled",
		ObservedGeneration: rr.Resource.Generation,
	}

	var allErrors error

	for i := range r.actions {
		if err := r.actions[i].Run(ctx, &rr); err != nil {
			allErrors = multierr.Append(allErrors, err)
		}
	}

	if allErrors != nil {
		reconcileCondition.Status = metav1.ConditionFalse
		reconcileCondition.Reason = "Failure"
		reconcileCondition.Message = "Failure"

		rr.Resource.Status.Phase = DaprPhaseError
	} else {
		rr.Resource.Status.ObservedGeneration = rr.Resource.Generation
		rr.Resource.Status.Phase = DaprPhaseReady
	}

	meta.SetStatusCondition(&rr.Resource.Status.Conditions, reconcileCondition)

	sort.SliceStable(rr.Resource.Status.Conditions, func(i, j int) bool {
		return rr.Resource.Status.Conditions[i].Type < rr.Resource.Status.Conditions[j].Type
	})

	//
	// Update status
	//

	err = r.Status().Update(ctx, rr.Resource)
	if err != nil && k8serrors.IsConflict(err) {
		l.Info(err.Error())
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		allErrors = multierr.Append(allErrors, err)
	}

	return ctrl.Result{}, allErrors
}
