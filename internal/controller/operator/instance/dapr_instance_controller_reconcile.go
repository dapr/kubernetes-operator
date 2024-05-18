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
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/reconciler"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/conditions"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	daprApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
)

func (r *Reconciler) Reconcile(ctx context.Context, res *daprApi.DaprInstance) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	rr := ReconciliationRequest{
		Client: r.Client(),
		NamespacedName: types.NamespacedName{
			Name:      res.Name,
			Namespace: res.Namespace,
		},
		ClusterType: r.ClusterType,
		Reconciler:  r,
		Resource:    res,
		Chart:       r.c,
		Overrides: map[string]interface{}{
			"dapr_operator":  map[string]interface{}{"runAsNonRoot": "true"},
			"dapr_placement": map[string]interface{}{"runAsNonRoot": "true"},
			"dapr_sentry":    map[string]interface{}{"runAsNonRoot": "true"},
			"dapr_dashboard": map[string]interface{}{"runAsNonRoot": "true"},
		},
	}

	l.Info("Reconciling", "resource", rr.NamespacedName.String())

	// by default, the controller expect the DaprInstance resource to be created
	// in the same namespace where it runs, if not fallback to the default namespace
	// dapr-system
	ns := os.Getenv(controller.NamespaceEnv)
	if ns == "" {
		ns = controller.NamespaceDefault
	}

	if res.Name != DaprInstanceResourceName || res.Namespace != ns {
		rr.Resource.Status.Phase = conditions.TypeError

		meta.SetStatusCondition(&rr.Resource.Status.Conditions, metav1.Condition{
			Type:   conditions.TypeReconciled,
			Status: metav1.ConditionFalse,
			Reason: conditions.ReasonUnsupportedConfiguration,
			Message: fmt.Sprintf(
				"Unsupported resource, the operator handles a single DaprInstance resource named %s in namespace %s",
				DaprInstanceResourceName,
				ns),
		})

		err := r.Client().Status().Update(ctx, rr.Resource)

		if err != nil && k8serrors.IsConflict(err) {
			l.Info(err.Error())
			return ctrl.Result{Requeue: true}, nil
		}

		return ctrl.Result{}, fmt.Errorf("error updating DaprInstance resource: %w", err)
	}

	//nolint:wrapcheck
	if rr.Resource.ObjectMeta.DeletionTimestamp.IsZero() {
		err := reconciler.AddFinalizer(ctx, r.Client(), rr.Resource, DaprInstanceFinalizerName)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		// Cleanup leftovers if needed
		for i := len(r.actions) - 1; i >= 0; i-- {
			if err := r.actions[i].Cleanup(ctx, &rr); err != nil {
				return ctrl.Result{}, err
			}
		}

		err := reconciler.RemoveFinalizer(ctx, r.Client(), rr.Resource, DaprInstanceFinalizerName)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	//
	// Reconcile
	//

	reconcileCondition := metav1.Condition{
		Type:               conditions.TypeReconciled,
		Status:             metav1.ConditionTrue,
		Reason:             conditions.ReasonReconciled,
		Message:            conditions.ReasonReconciled,
		ObservedGeneration: rr.Resource.Generation,
	}

	errs := make([]error, 0, len(r.actions)+1)

	for i := range r.actions {
		if err := r.actions[i].Run(ctx, &rr); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		reconcileCondition.Status = metav1.ConditionFalse
		reconcileCondition.Reason = conditions.ReasonFailure
		reconcileCondition.Message = conditions.ReasonFailure

		rr.Resource.Status.Phase = conditions.TypeError
	} else {
		rr.Resource.Status.ObservedGeneration = rr.Resource.Generation
		rr.Resource.Status.Phase = conditions.TypeReady
	}

	meta.SetStatusCondition(&rr.Resource.Status.Conditions, reconcileCondition)

	sort.SliceStable(rr.Resource.Status.Conditions, func(i, j int) bool {
		return rr.Resource.Status.Conditions[i].Type < rr.Resource.Status.Conditions[j].Type
	})

	//
	// Update status
	//

	err := r.Client().Status().Update(ctx, rr.Resource)

	if err != nil && k8serrors.IsConflict(err) {
		l.Info(err.Error())
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		errs = append(errs, err)
	}

	return ctrl.Result{}, errors.Join(errs...)
}
