package reconciler

import (
	"context"
	"fmt"
	"reflect"

	"github.com/dapr/kubernetes-operator/pkg/conditions"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/dapr/kubernetes-operator/pkg/controller"
	"github.com/dapr/kubernetes-operator/pkg/controller/client"
)

type Reconciler interface {
	Client() *client.Client
}

type Action[T any] interface {
	Configure(ctx context.Context, c *client.Client, b *builder.Builder) (*builder.Builder, error)
	Run(ctx context.Context, rc *T) error
	Cleanup(ctx context.Context, rc *T) error
}

func AddFinalizer(ctx context.Context, client ctrlClient.Client, o ctrlClient.Object, name string) error {
	if !ctrlutil.AddFinalizer(o, name) {
		return nil
	}

	err := client.Update(ctx, o)
	if k8serrors.IsConflict(err) {
		return fmt.Errorf("conflict when adding finalizer to %s/%s: %w", o.GetNamespace(), o.GetName(), err)
	}

	if err != nil {
		return fmt.Errorf("failure adding finalizer to %s/%s: %w", o.GetNamespace(), o.GetName(), err)
	}

	return nil
}

func RemoveFinalizer(ctx context.Context, client ctrlClient.Client, o ctrlClient.Object, name string) error {
	if !ctrlutil.RemoveFinalizer(o, name) {
		return nil
	}

	err := client.Update(ctx, o)
	if k8serrors.IsConflict(err) {
		return fmt.Errorf("conflict when removing finalizer to %s/%s: %w", o.GetNamespace(), o.GetName(), err)
	}

	if err != nil {
		return fmt.Errorf("failure removing finalizer from %s/%s: %w", o.GetNamespace(), o.GetName(), err)
	}

	return nil
}

type BaseReconciler[T controller.ResourceObject] struct {
	Log             logr.Logger
	Name            string
	Namespace       string
	FinalizerName   string
	FinalizerAction func(ctx context.Context, res T) error
	Delegate        reconcile.ObjectReconciler[T]
	Client          ctrlClient.Client
}

//nolint:forcetypeassert,wrapcheck,nestif
func (s *BaseReconciler[T]) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	res := reflect.New(reflect.TypeOf(*new(T)).Elem()).Interface().(T)
	if err := s.Client.Get(ctx, req.NamespacedName, res); err != nil {
		return ctrl.Result{}, ctrlClient.IgnoreNotFound(err)
	}

	if res.GetName() != s.Name || res.GetNamespace() != s.Namespace {
		res.GetStatus().Phase = conditions.TypeError

		meta.SetStatusCondition(&res.GetStatus().Conditions, metav1.Condition{
			Type:   conditions.TypeReconciled,
			Status: metav1.ConditionFalse,
			Reason: conditions.ReasonUnsupportedConfiguration,
			Message: fmt.Sprintf(
				"Unsupported resource, the operator handles a single %s resource named %s in namespace %s",
				res.GetObjectKind().GroupVersionKind().String(),
				s.Name,
				s.Namespace),
		})

		err := s.Client.Status().Update(ctx, res)

		if err != nil && k8serrors.IsConflict(err) {
			s.Log.Info(err.Error())
			return ctrl.Result{Requeue: true}, nil
		}

		if err != nil {
			return ctrl.Result{}, fmt.Errorf("error updating %s resource: %w", res.GetObjectKind().GroupVersionKind().String(), err)
		}

		return ctrl.Result{}, nil
	}

	//nolint:wrapcheck
	if res.GetDeletionTimestamp().IsZero() {
		err := AddFinalizer(ctx, s.Client, res, s.FinalizerName)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if s.FinalizerAction != nil {
			err := s.FinalizerAction(ctx, res)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		err := RemoveFinalizer(ctx, s.Client, res, s.FinalizerName)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	return s.Delegate.Reconcile(ctx, res)
}
