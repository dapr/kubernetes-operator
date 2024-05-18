package reconciler

import (
	"context"
	"fmt"

	"github.com/dapr-sandbox/dapr-kubernetes-operator/pkg/controller/client"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
