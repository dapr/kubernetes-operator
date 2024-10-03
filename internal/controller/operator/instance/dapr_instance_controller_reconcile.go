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
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/dapr/kubernetes-operator/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dapr/kubernetes-operator/pkg/conditions"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	daprApi "github.com/dapr/kubernetes-operator/api/operator/v1beta1"
)

func (r *Reconciler) reconciliationRequest(res *daprApi.DaprInstance) (ReconciliationRequest, error) {
	rr := ReconciliationRequest{
		Client:     r.Client(),
		Reconciler: r,
		Resource:   res,
		Helm: Helm{
			engine:      r.helmEngine,
			chartDir:    r.helmOptions.ChartsDir,
			ChartValues: make(map[string]interface{}),
			chartOverrides: map[string]interface{}{
				"dapr_operator":         map[string]interface{}{"runAsNonRoot": "true"},
				"dapr_placement":        map[string]interface{}{"runAsNonRoot": "true"},
				"dapr_sentry":           map[string]interface{}{"runAsNonRoot": "true"},
				"dapr_dashboard":        map[string]interface{}{"runAsNonRoot": "true"},
				"dapr_sidecar_injector": map[string]interface{}{"runAsNonRoot": true},
			},
		},
	}

	if res.Spec.Values != nil {
		if err := json.Unmarshal(res.Spec.Values.RawMessage, &rr.Helm.ChartValues); err != nil {
			return ReconciliationRequest{}, fmt.Errorf("unable to decode chart values: %w", err)
		}
	}

	return rr, nil
}

func (r *Reconciler) Reconcile(ctx context.Context, res *daprApi.DaprInstance) (ctrl.Result, error) {
	rr, err := r.reconciliationRequest(res)
	if err != nil {
		return ctrl.Result{}, err
	}

	_, err = rr.Chart(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	l := log.FromContext(ctx)
	l.Info("Reconciling", "resource", rr.Resource.Name)

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

	err = r.Client().ApplyStatus(
		ctx,
		rr.Resource,
		client.ForceOwnership,
		client.FieldOwner(controller.FieldManager),
	)

	if err != nil {
		errs = append(errs, err)
	}

	return ctrl.Result{}, errors.Join(errs...)
}

func (r *Reconciler) Cleanup(ctx context.Context, res *daprApi.DaprInstance) error {
	rr, err := r.reconciliationRequest(res)
	if err != nil {
		return err
	}

	_, err = rr.Chart(ctx)
	if err != nil {
		return err
	}

	l := log.FromContext(ctx)
	l.Info("Cleanup", "resource", rr.Resource.Name)

	// Cleanup leftovers if needed
	for i := len(r.actions) - 1; i >= 0; i-- {
		if err := r.actions[i].Cleanup(ctx, &rr); err != nil {
			return fmt.Errorf("failure running cleanup action: %w", err)
		}
	}

	return nil
}
